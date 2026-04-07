---
applyTo: "**"
---

# 技术选型规则

## 当前已知技术栈

- 后端语言：Go 1.26.1，模块路径 `github.com/lpxxn/blink`
- 启动入口：`cmd/main.go`
- HTTP 路由：[Gin](https://github.com/gin-gonic/gin)；OpenAPI 代码生成见 `docs/oapi-codegen.md`
- 协议目录：`infrastructure/interface/http/` 与 `infrastructure/interface/grpc/`

## 明确偏好

- 数据库：先用 SQLite 落地，保留切换到 MySQL 或 PostgreSQL 的能力。
- 日志：优先使用 `zap`，结构化、可检索、适合生产排障。
- API：同时支持 HTTP RESTful API 与 GraphQL。
- 测试：单元测试优先轻量，集成测试可使用 Docker 环境。

## 技术决策原则

- 优先使用 Go 标准库，避免过早引入重量级框架。
- 引入第三方库前，先确认是否真正能降低复杂度。
- 新依赖必须与 DDD 分层兼容，不能把框架侵入领域层。
- 只在明确需要时引入 HTTP、gRPC、GraphQL、消息队列、缓存等组件。

## 后端实现建议

- Web 层优先保持轻量，路由、中间件和 handler 尽量薄。
- 配置、日志、数据库连接、客户端封装应集中在基础设施层。
- 如需 ORM，应限制在 `infrastructure/persistence/`，不要泄漏到领域层。
- 如需 GraphQL，应将 schema/resolver 视作接口层实现，而不是领域层。
- 配置管理应能支持不同运行环境和多数据库驱动切换。

## 依赖管理

- Go 依赖统一通过 `go.mod` / `go.sum` 管理。
- 不引入多个功能重叠的库。
- 依赖升级应优先保证兼容性与安全性。
- 对基础库优先做少而稳的选择（日志、数据库访问、配置、HTTP 各保留一套主实现）。

## 不推荐做法

- 在项目早期一次性引入过多中间件。
- 因为"以后可能会用到"而提前接入复杂基础设施。
- 让某个具体库的类型或错误模型渗透整个业务层。
