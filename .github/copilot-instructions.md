# Blink 项目总览

## 项目定位

Blink 是一个轻量、高性能的微博客平台，目标是支持用户发布短内容、日常动态和视觉内容。

## 产品范围

- 用户注册、登录、身份认证与权限控制
- 发布短文、想法、图片、视频等内容
- 内容查询、时间线、详情页、用户主页等读取场景
- 内容审核与敏感内容处理
- 面向不同客户端提供 REST 与 GraphQL 两套接口

## 技术方向

- 后端：Go（模块路径 `github.com/lpxxn/blink`，入口 `cmd/main.go`）
- 架构：DDD 分层（`domain/` → `application/` → `infrastructure/` → `cmd/`）
- 数据库：SQLite 先行，保留切换 MySQL / PostgreSQL 的能力
- 日志：`zap`，结构化、可检索
- API：HTTP RESTful（OpenAPI + oapi-codegen）+ GraphQL

## 规则优先级

详细规则分布在 `.github/instructions/` 下，按以下顺序应用：

1. `01-architecture` — 判断代码落在哪一层
2. `02-tech-stack` — 技术选型约束
3. `03-code-style` — 代码规范
4. `04-database` — 数据库与迁移规则
5. `05-api` — REST / GraphQL / gRPC 规则
6. `06-testing` — 测试策略
7. `07-observability` — 日志、指标、追踪
8. `08-security` — 安全规则
9. `09-performance` — 性能规则
10. `10-devops` — 部署与运维

## 基本原则

- 以 DDD 分层为中心，避免跨层耦合。
- 先定义清晰的领域概念，再补充接口、存储和交付层。
- 所有新增实现都应与现有骨架保持一致，不要绕过目录边界。
- 不要因为当前代码少，就把所有逻辑直接写进 `cmd/main.go`。
- 不要在 `infrastructure/` 中定义业务规则。
- 不要让 DTO、数据库模型、HTTP 请求对象直接污染领域对象。
