---
applyTo: "**/*.go"
---

# DDD 架构规则

## 分层目标

采用以 DDD 为核心的分层组织方式，新增代码必须遵循以下边界：

- `domain/`：表达核心业务语义，不依赖传输协议、数据库细节或框架细节。
- `application/`：编排用例，连接领域对象与外部能力。
- `infrastructure/`：实现技术细节，例如 HTTP、gRPC、数据库、缓存、外部服务适配。
- `cmd/`：组装依赖、启动服务、加载配置，不写业务逻辑。

## 依赖方向

只允许稳定依赖指向内部核心：

```
cmd -> infrastructure -> application -> domain
```

- `domain/` 不依赖 `application/`、`infrastructure/`、`cmd/`。
- `application/` 不依赖具体协议和具体数据库实现。
- `infrastructure/` 可以依赖 `application/` 和 `domain/`，用于承接适配器实现。
- `cmd/` 只做启动与装配，不写业务逻辑。

## 目录职责

### `domain/`

承载：实体（Entities）、值对象（Value Objects）、聚合根（Aggregates）、领域服务（Domain Services）、仓储接口（Repository Interfaces）、领域事件（Event）

禁止放入：HTTP 请求/响应结构、ORM 模型、SQL 语句、日志框架直接调用

### `application/`

承载：Use Case / Command / Query Handler、DTO、应用服务、事务边界协调、权限校验编排

要求：面向领域模型和仓储接口编程，不重新实现领域规则。

### `infrastructure/`

已预留子目录：
- `interface/http/`：HTTP Handler、路由绑定、中间件
- `interface/grpc/`：gRPC Handler、proto 对接逻辑
- `repository/`：仓储接口的具体实现
- `persistence/`：数据库模型、查询封装、连接初始化
- `adapter/`：调用外部系统的网关、SDK 包装器

## 设计规则

- 用例先行：先从应用层的行为设计功能，再回推领域对象与基础设施实现。
- 领域隔离：领域层应可在脱离 Web/DB 环境下被测试。
- 接口分离：在核心层定义抽象，在基础设施层实现。
- 数据转换显式化：DTO、领域对象、持久化模型之间必须有明确转换。

## 命名建议

- Use case 以动作命名，如 `CreatePost`, `PublishBlink`, `GetTimeline`。
- 仓储接口以语义命名，如 `PostRepository`, `UserRepository`。
- 基础设施实现可体现技术细节，如 `MySQLPostRepository`, `HTTPPostHandler`。

## 新功能默认流程

1. 领域概念是什么？
2. 用例输入输出是什么？
3. 需要哪些仓储/外部接口？
4. HTTP/gRPC/API 层如何映射？
5. 数据库存储如何建模？

## 反模式

- 在 Handler 中直接写核心业务。
- 在 Repository 中拼接完整业务流程。
- 让 ORM 模型充当领域模型。
- 让一个包同时承担"协议 + 业务 + 存储"三种职责。
