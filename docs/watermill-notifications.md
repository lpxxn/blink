# Blink 站内通知：Watermill + Redis Stream 说明

本文说明本项目中 [Watermill](https://watermill.io/) 的用途、数据流、配置与运维注意事项。适合尚未使用过 Watermill 的读者。

整体分层与模块关系见 [architecture.md](./architecture.md) 第 4 节序列图。

---

## 1. Watermill 是什么？在本项目里做什么？

**Watermill** 是 Go 的**消息基础设施库**：统一抽象「发布消息 / 订阅消息 / 路由与中间件」，底层可对接 Redis Stream、Kafka、RabbitMQ 等。

在本仓库中，Watermill **只用于「站内通知」这一条业务链**：

1. HTTP 或应用服务在业务成功后，不直接写 `notifications` 表，而是把**一条 JSON 事件**发到 **Redis Stream**。
2. 同进程（或将来独立进程）里运行的 **Watermill `message.Router`** 从 Stream 读出消息，调用 `application/notification.Service` 的 `OnNewReply` / `OnPostRemoved` 等方法，**再写入数据库**。

这样做的效果：

- **发布端**（API、Admin）只依赖接口 `application/eventing.NotificationPublisher`，不关心 Redis 细节。
- **消费端**与发布解耦，可单独扩缩容、重试（由 Redis 消费者组与 Watermill 行为共同决定）。
- **领域约定**：JSON 字段 `type` 的取值来自 `domain/event/notification.go`，发布与消费均引用同一常量，避免魔法字符串漂移。`domain/event/events.go` 中为同一业务语义保留了**点分式事件名**（如 `reply.posted`）及 **Payload** 结构体，便于将来与其它消费者或统一 envelope 对齐；当前 Watermill 路径仍以 `notification.go` 为准，演进时需同时改 Publisher、`dispatchNotificationEvent` 与本文协议表。

---

## 2. 核心概念速查（对照本项目）

| 概念 | 含义 | 在本项目中的对应 |
|------|------|------------------|
| **Publisher** | 把 `message.Message` 发到某个 **topic** | `redisstream.Publisher` → topic 即 **Stream 名称** |
| **Subscriber** | 从 topic 订阅消息流 | `redisstream.Subscriber` + **Consumer Group** |
| **Topic** | 逻辑通道名 | 常量 `blink.notification.events`（`infrastructure/messaging/notification_topic.go`） |
| **message.Message** | 一条消息：UUID + `Payload`（[]byte）+ Metadata | `Payload` 为本项目自定义的 **JSON** |
| **Router** | 注册「从哪个 Subscriber 读、用哪个函数处理」 | `message.NewRouter` + `AddConsumerHandler` |
| **Handler 返回值** | `error == nil` → **Ack**；非 nil → **Nack**（可能重投，视实现而定） | 见第 6 节错误与重试 |

官方文档入口：<https://watermill.io/docs/>；Redis Stream 组件：<https://watermill.io/pubsubs/redisstream/>。

---

## 3. 端到端数据流（从发评论到用户看到通知）

```
用户 POST /api/posts/:id/replies
    → replies 落库成功
    → httpapi.Server.NotifyEvents（NotificationWatermillPublisher）
    → Watermill Publisher → Redis XADD stream "blink.notification.events"
    → Watermill Subscriber（消费者组 blink-notify）读取
    → Router Handler → dispatchNotificationEvent
    → notification.Service.OnNewReply / OnReplyToYourComment
    → GORM 写入 notifications 表
用户 GET /api/me/notifications
    → 直接读 DB（不经过 Watermill）
```

**管理员下架 / 申诉裁决** 类似：`application/admin.Service` 在 `NotifyEvents` 上调用 `PublishPostRemoved` / `PublishAppealResolved`。

---

## 4. Redis Stream 与「消费者组」要注意什么？

- **Stream 名称**：`blink.notification.events`（与代码中 `TopicNotificationEvents` 一致）。
- **消费者组**（默认 `blink-notify`）：同一组内**每条消息只会被组内一个消费者处理**；适合多副本抢消费、断线重连后 **Pending** 再投递。
- **Consumer 名称**（实例标识）：必须**在同一消费者组内尽量唯一**。本仓库默认 `主机名-PID`；多容器部署请显式设置 `BLINK_WATERMILL_CONSUMER`，避免两实例同名导致行为怪异。
- **`OldestId`（`BLINK_NOTIFICATION_STREAM_FROM`）**：
  - 默认 **`$`**：消费者组**首次创建**时，通常只消费**创建组之后**新写入的消息（具体行为以 Redis / 库版本与建组方式为准，部署前建议在测试环境验证）。
  - 若需要「从头回放」历史消息，可改为 **`0`** 等（注意可能一次性灌入大量积压，谨慎使用）。

---

## 5. 消息 JSON 协议（与代码严格一致）

每条消息的 `Payload` 为 **UTF-8 JSON**。必须包含字段 **`type`**（字符串），取值来自 `domain/event/notification.go`：

| `type` 值 | 含义 | 主要字段 |
|-----------|------|----------|
| `reply_to_post` | 帖子有新评论（通知楼主） | `post_author_id`, `post_id`, `reply_id`, `snippet`（均为 JSON；**整型 ID 使用 `json:",string"` 序列化为字符串**，避免 JS 精度问题） |
| `reply_to_comment` | 有人回复了你的评论 | `parent_author_id`, `post_id`, `reply_id`, `snippet` |
| `post_removed` | 帖子被管理员下架 | `author_id`, `post_id`, `reason` |
| `appeal_resolved` | 申诉/复核已裁决 | `author_id`, `post_id`, `approved`, `admin_note` |

发布端实现：`infrastructure/messaging/notification_watermill_publisher.go`（`type` 使用 `domain/event` 包内常量）。  
消费端解析：`infrastructure/messaging/notification_watermill_consumer.go` 中 `dispatchNotificationEvent`。

**与 `domain/event/events.go` 的关系**：该文件中的点分名与 `*Payload` 与 `notification.go` 的 `type` 字符串**语义对应但字面值不同**；新增或变更事件时，应同时更新两处（或完成统一重构），并跑通发布/消费与本文档表格。

---

## 6. 错误处理、Ack / Nack 与「毒消息」

Handler 逻辑（简化）：

- **`json.Unmarshal` 失败或 `type` 无法识别**：返回 **`nil`**（视为成功）→ 消息被 **Ack**，**不会**无限重试。原因是：这类数据无法修复，重试无意义，属于**毒消息丢弃策略**（当前未写入死信队列）。
- **`notification.Service` 返回错误**（例如 DB 失败）：返回 **该 error** → Watermill/Redis 侧按实现 **Nack 或重投**，便于**短暂故障**恢复。

运维建议：若长期出现 DB 错误，应查 DB/连接；若出现大量畸形消息，应查发布端版本是否一致。

---

## 7. 环境变量一览

| 变量 | 作用 | 默认 |
|------|------|------|
| `BLINK_REDIS_ADDR` | Redis 地址（与 Session 等共用客户端） | `127.0.0.1:6379` |
| `BLINK_DISABLE_NOTIFICATION_CONSUMER` | 任意非空：**不启动**本进程内的 Watermill Router 消费者 | 空（即**启用**消费者） |
| `BLINK_WATERMILL_CONSUMER` | Redis 消费者组内的 Consumer 名 | `hostname-pid` |
| `BLINK_WATERMILL_CONSUMER_GROUP` | 消费者组名 | `blink-notify` |
| `BLINK_NOTIFICATION_STREAM_FROM` | 建组/读取起点（如 `$`、`0`） | `$` |

**禁用消费者、仍发布事件的场景**：例如通知消费迁到**独立 worker 进程**，主 API 只 `Publish`；此时必须保证**别处**有进程跑同一套 `Subscriber` + `RunNotificationWatermillRouter`，否则消息只堆在 Stream 里，用户永远收不到站内信。

---

## 8. 进程内启动顺序（`cmd/main.go`）

1. 连接 Redis（与现有 Session 等共用 `*redis.Client`）。
2. 构造 `redisstream.NewPublisher`，再包一层 `messaging.NewNotificationWatermillPublisher` → 注入 `httpapi.Server.NotifyEvents`、`admin.Service.NotifyEvents`。
3. 若未设置 `BLINK_DISABLE_NOTIFICATION_CONSUMER`：
   - `redisstream.NewSubscriber`（配置 ConsumerGroup / Consumer / OldestId 等）；
   - `messaging.RunNotificationWatermillRouter`：内部 `go router.Run(ctx)`，并 **`<-router.Running()`** 等待路由就绪后再继续启动 Gin（避免刚上线时短时间消费不到消息）。
4. `defer` 顺序：`Router.Close` → `Subscriber.Close` → `Publisher.Close`（后注册的先执行）。

`RunNotificationWatermillRouter` 使用的 `ctx` 当前为 **`context.Background()`**，即随进程生命周期；优雅停机若要停止 Router，需要改为可取消的 `Context` 并在收到信号时 `cancel`（当前未实现，属改进项）。

---

## 9. 本地调试建议

1. 启动 Redis，保证 `BLINK_REDIS_ADDR` 可连。
2. 跑主程序，发一条会触发通知的请求（例如给帖子评论）。
3. 用 `redis-cli` 查看 Stream：
   - `XLEN blink.notification.events`
   - `XRANGE blink.notification.events - + COUNT 5`
4. 查消费者组：
   - `XINFO GROUPS blink.notification.events`
   - `XPENDING blink.notification.events blink-notify`

若 `XLEN` 增加但 DB 里没有新通知，检查消费者是否启动、组名/Consumer 是否一致、日志里是否有 `notification event handler:` 或 `watermill router stopped:`。

---

## 10. 依赖与版本

- `github.com/ThreeDotsLabs/watermill`（Router、`message` 抽象）
- `github.com/ThreeDotsLabs/watermill-redisstream`（Redis Stream 的 Publisher/Subscriber）
- `github.com/redis/go-redis/v9`（与项目其余 Redis 用法一致）

版本以 `go.mod` 为准；升级 Watermill 或 redisstream 时务必跑全量测试，并重点回归 **JSON 字段 tag**（`,string`）是否与反序列化一致。

---

## 11. 常见问题（FAQ）

**Q：消息会「恰好一次」投递吗？**  
A：Redis Stream + 消费者组通常是**至少一次**；业务上 `notifications` 若需幂等，应在落库侧用业务键去重（当前实现未做，重复消费可能产生重复通知行）。

**Q：能和 Kafka 混用吗？**  
A：可以另建 Publisher/Subscriber 实现同一 `eventing.NotificationPublisher` 接口；本仓库目前仅 Redis。

**Q：为什么 Handler 里用 `context.Background()` 而不是请求的 Context？**  
A：消费发生在请求结束之后，请求 ctx 往往已取消；若需要链路追踪，应传入派生的后台 ctx 或把 trace id 放进 Watermill `message.Metadata`（当前未实现）。

**Q：发布失败会怎样？**  
A：`Publish` 返回 error 时，当前调用方多为 `_ = ...` 忽略，**用户操作已成功但可能无通知**；重要业务可改为记录日志或指标（改进项）。

---

## 12. 相关源码索引

| 文件 | 职责 |
|------|------|
| `application/eventing/notification_publisher.go` | 发布接口定义 |
| `application/eventing/nop_publisher.go` | 空实现（测试可用） |
| `domain/event/notification.go` | Stream JSON 中 `type` 字符串常量（发布/消费共用） |
| `domain/event/events.go` | 点分式事件名与 Payload 结构（与 `notification.go` 语义对应，供演进统一） |
| `infrastructure/messaging/notification_topic.go` | Stream 名常量 |
| `infrastructure/messaging/notification_watermill_publisher.go` | JSON 序列化 + `message.Publish` |
| `infrastructure/messaging/notification_watermill_consumer.go` | Router、反序列化、`dispatchNotificationEvent` |
| `cmd/main.go` | 组装 Publisher/Subscriber/Router、环境变量 |
| `infrastructure/interface/http/api/replies.go` | 评论后 `NotifyEvents` 发布 |
| `application/admin/service.go` | 下架/申诉结果 `NotifyEvents` 发布 |

---

## 13. 小结

- Watermill 在本项目 = **Redis Stream 上的轻量消息总线**，专门承载「站内通知」异步落库。
- 务必理解 **Consumer Group / Consumer 名 / OldestId** 对行为的影响。
- 协议以 **`type`（`domain/event/notification.go`）+ JSON 字段** 为准，升级时注意与 Publisher、`dispatchNotificationEvent` 及 `events.go` 语义约定同步。
- 需要更严格的集成或「恰好一次」，可在本设计之上增量加 **Outbox、幂等键、死信队列** 等，而不必换掉 Watermill 核心用法。
