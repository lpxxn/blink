# Flutter 客户端计划（Blink）

本项目将新增 Flutter 客户端（iOS/Android 优先，可扩展 Web/Desktop），对接 Blink 后端（OpenAPI、OAuth2、Cookie Session、图片上传、通知等）。本文档覆盖技术选型、架构分层、工程化规范、Riverpod（代码生成模式）落地方式，以及里程碑计划。

## 目标与非目标

- **目标**
  - iOS/Android 双端一致体验，工程结构可长期迭代
  - 与后端契约（`api/openapi/openapi.yaml`）对齐，减少联调成本
  - 状态管理统一使用 **Riverpod 最新 codegen 模式**（`@riverpod` + generator）
  - 可测试、可观测（日志/崩溃/性能）、多环境可控（dev/staging/prod）
- **非目标（第一阶段不做或弱化）**
  - 复杂离线优先（可先做基础缓存，离线完整能力后续迭代）
  - 过度通用的组件库（先沉淀项目内常用组件与规范）

## 技术选型（推荐组合）

### Flutter & Dart

- Flutter：稳定版（建议每 1–2 个小版本评估升级）
- Dart：随 Flutter
- 工程规范：`flutter_lints` + 项目自定义 `analysis_options.yaml`

### 状态管理（强制）

- `flutter_riverpod`
- `riverpod_annotation`
- `riverpod_generator`
- `build_runner`

> 约束：业务状态 **只允许** 使用 `@riverpod` 生成 provider；禁止手写 `Provider(...)` 作为业务入口（基础设施 provider 例外，例如 `dioProvider`）。

### 路由

- `go_router`：支持深链、嵌套路由、重定向/守卫
- 路由重定向由 Riverpod 的登录态/权限态驱动（避免在 UI 到处写判断）

### 网络、契约与序列化

- HTTP：`dio`
- API 封装：优先 `retrofit`（搭配 Dio）
- JSON：`json_serializable`
- 不可变/联合类型：`freezed`

与后端契约的对齐方式：

- **接口契约来源**：`api/openapi/openapi.yaml`
- **Snowflake ID 约定**：JSON 中 ID 使用 **字符串**（后端已有该约定，客户端按字符串处理）

### 本地存储

- 配置：`shared_preferences`
- 敏感信息：`flutter_secure_storage`（例如 token/refresh 信息，若后端为 Cookie Session，则主要存服务端会话标识与 CSRF/设备标识等）
- 结构化存储（需要再引入）：`drift` 或 `isar`

### 可观测性

- 日志：`logger`
- 崩溃与性能：`sentry_flutter`（或 Firebase Crashlytics/Performance，二选一）
- 埋点：抽象 `AnalyticsService` 接口，通过 provider 注入实现

## 与 Blink 后端对接要点

### 认证与会话（对齐 README）

后端支持：

- 邮箱密码注册：`POST /auth/register`
- OAuth2 授权码登录（第三方或 builtin IdP）
- 会话 Cookie：`blink_session`（Redis 存储会话）

客户端策略建议：

- HTTP 请求必须携带 Cookie（Dio 配置 Cookie 管理；移动端需要 CookieJar 支持）
- 登录态以“是否存在有效会话”为准：提供 `SessionRepository`（读取/写入会话相关本地信息）
- `go_router` 通过 `authStateProvider` 做登录重定向

### 图片上传

- 接口：`POST /api/uploads`（multipart 字段 `file`）
- 默认上传目录：`data/uploads`，通过 `/uploads/...` 访问
- 客户端：统一封装 `UploadRepository`（包含压缩/失败重试/进度上报策略）

### 通知

后端站内通知为异步写入（Watermill + Redis Stream）。客户端可按：

- 第一阶段：轮询拉取通知列表/未读数（接口若现有或后续补齐）
- 第二阶段：若后端提供 SSE/WebSocket，再升级为实时

## 架构与目录结构（建议）

> 建议将 Flutter 客户端放在仓库内 `client/flutter/`（或 `apps/flutter/`），并在其内自成一套工程与 CI。

推荐目录（以 `lib/` 为根）：

- `app/`：启动与组装（router、theme、env、providers）
- `core/`：跨业务能力（network、error、logging、storage、widgets）
- `features/`：按业务模块拆分（auth、feed、post、profile、notifications...）
  - `data/`：dto、api client、datasource、repository impl
  - `domain/`：entity、repository interface（可选）
  - `application/`：usecase/服务编排（推荐）
  - `presentation/`：pages/widgets/providers（页面相关）

依赖方向：

- `presentation -> application -> domain -> data`
- 通过 Riverpod provider 进行依赖注入与组装，避免层间直接 new 具体实现。

## Riverpod（codegen）落地规范（重点）

### Provider 类型使用原则

- **读取型异步数据**：用 `@riverpod` 返回 `Future<T>` / `Stream<T>`
  - UI 消费 `AsyncValue<T>`：`loading/error/data`
- **可变同步状态**：用 `@riverpod class Xxx extends _$Xxx`（Notifier）
  - 适合表单字段、筛选条件、简单 UI state
- **带请求生命周期的可变状态**：用 `AsyncNotifier`
  - 适合登录、提交、分页、刷新等

### 统一错误模型

- Data 层将 Dio/解析/业务错误统一映射为 `AppError`
- Presentation 层只负责把 `AppError` 映射为文案与交互（toast/dialog/empty state）

### 副作用约束

- 导航、toast、弹窗等副作用放在 `ref.listen(...)` 或 `ProviderListener`，禁止在 `build` 内触发。

### 代码生成命令

- 一次性生成：
  - `dart run build_runner build -d`
- 开发 watch：
  - `dart run build_runner watch -d`

CI 必须包含生成校验（避免本地忘记生成导致 CI/打包失败）。

## 网络层设计规范

- `dioProvider` 产出配置完毕的 Dio（baseUrl、超时、拦截器、日志）
- 拦截器建议包含：
  - session/cookie 处理（如需要）
  - 401/未授权统一处理（驱动 `authStateProvider` 变更）
  - 错误映射为 `AppError`
- Repository 对外返回领域模型或 view model；避免把 DTO 直接暴露到 UI。

## 测试策略（最小可用）

- 单测：usecase、repository、mapper（不依赖 Flutter）
- widget test：页面在 provider overrides 下验证 loading/error/content
- 集成测试（后续）：登录 -> 浏览 feed -> 发帖 -> 上传图片 -> 评论

## 工程化与质量门禁（建议）

- `dart format`、`dart analyze`、`flutter test` 进 CI
- 依赖治理：定期 `flutter pub outdated`，按节奏升级
- 多环境：`--dart-define=ENV=staging` 等注入 baseUrl 与开关

## 本地运行（多环境）

在 `client/flutter/` 下通过 `--dart-define` 注入环境名与 API 基址（与 `lib/core/env/app_env_provider.dart` 一致）：

```bash
cd client/flutter
flutter run \
  --dart-define=ENV=dev \
  --dart-define=API_BASE_URL=http://127.0.0.1:11110
```

未指定时默认 `ENV=dev`、`API_BASE_URL=http://127.0.0.1:11110`。

## 里程碑计划

### 阶段 A：骨架（1–3 天）

- Flutter 工程初始化（建议路径 `client/flutter/`）
- Riverpod codegen 跑通（build_runner）
- go_router 路由骨架（登录页/主页占位）
- Dio + 错误体系 + 日志

### 阶段 B：认证闭环（3–7 天）

- 注册/登录（对齐 `/auth/register` 与 OAuth 流程）
- 会话维持与退出
- 路由守卫与跳转

### 阶段 C：核心业务（按模块迭代）

- Feed 列表、详情、评论
- 发帖、图片上传
- 通知列表/未读（轮询版）

### 阶段 D：上线准备（1–2 周）

- 性能/首屏优化
- 崩溃与埋点接入
- 测试补齐与灰度策略

