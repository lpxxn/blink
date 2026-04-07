---
applyTo: "infrastructure/interface/**,api/**"
---

# API 规则：REST + GraphQL + gRPC

## OpenAPI 与 oapi-codegen

- **规范位置**：`api/openapi/openapi.yaml`（禁止手改生成代码）。
- **生成代码**：`api/gen/apigen.gen.go`，通过 `go generate ./api/gen/...` 重新生成。
- **当前配置**：仅为带标签 `servergen` 的操作生成 Gin 注册代码，其余路径由手写 Gin 路由实现。
- 修改 YAML 后必须跑生成并使 `go build ./...` 通过后再提交。
- `ServerInterface` 由生成代码定义；具体业务在 `infrastructure/interface/http/` 实现，通过 `apigen.RegisterHandlers` 与生成路由对齐。

## 总体原则

- API 层属于接口层，不承载核心业务规则。
- 输入校验、鉴权、协议映射在接口层完成。
- 业务决策交给应用层/领域层。
- 错误响应要稳定、明确、可观测。
- 同一业务能力可同时暴露为 REST 和 GraphQL，但只能复用一套应用层用例。

## REST 设计规则

- 资源命名使用名词复数，如 `/users`, `/posts`, `/blinks`。
- 使用 HTTP 方法表达动作语义：`GET` 查询、`POST` 创建、`PUT/PATCH` 更新、`DELETE` 删除。
- URL 保持简洁，不把业务流程塞进路径。
- 列表接口必须支持分页、排序与必要过滤。

## REST 响应规则

- 成功响应结构应稳定。
- 错误响应应包含：机器可识别错误码、人类可读信息、必要时的追踪 ID。
- 不把内部错误堆栈直接暴露给客户端。
- 错误文案要对客户端友好，面向用户的提示与面向日志的诊断信息应分离。

## GraphQL 设计规则

- GraphQL 仅作为另一种接口层，不替代领域模型。
- Resolver 只做参数解析、鉴权和用例调用。
- Schema 命名应贴近领域语言，而非数据库字段名。
- 谨慎处理深层嵌套和批量查询，避免性能失控。
- 优先解决 N+1 查询问题，再扩展复杂查询能力。

## REST 与 GraphQL 的关系

- 不要在两套 API 中复制两份业务逻辑，应统一调用应用层用例。
- REST 更适合标准 CRUD、后台管理和简单集成。
- GraphQL 更适合前端按需取数和复杂聚合查询。

## 版本管理

- 对外接口一旦发布，应有兼容策略。
- 破坏性变更必须通过版本升级、字段弃用周期或明确迁移方案处理。

## gRPC 约束

- protobuf 只是传输契约，不是领域模型。
- gRPC handler 负责协议转换，不直接操作数据库。
- 公共业务仍走应用层。
