# oapi-codegen v2 在本仓库中的用法

本文说明 [oapi-codegen v2](https://github.com/oapi-codegen/oapi-codegen) 的**命令行参数**、**YAML 配置**与本项目路径的对应关系。生成器版本与 `api/gen/doc.go` 中 `go:generate` 引用的标签一致（当前为 **v2.6.0**）。

## 本仓库中的文件

| 路径 | 作用 |
|------|------|
| `api/openapi/openapi.yaml` | OpenAPI 3.x 规范入口（单一事实来源；**全量 HTTP 文档**） |
| `api/openapi/oapi-codegen.yaml` | 生成器配置；当前含 `output-options.include-tags: [servergen]`，只为带 **`servergen`** 标签的操作生成 Gin 代码 |
| `api/gen/doc.go` | `go:generate` 指令，调用上述配置与规范 |
| `api/gen/apigen.gen.go` | **生成物，禁止手改** |

新增需 **参与代码生成** 的 operation 时，在 `openapi.yaml` 里为该 operation 增加标签 **`servergen`**（可与 `system` 等并存）。

在仓库根目录执行：

```bash
go generate ./api/gen/...
```

或在 `api/gen` 目录执行 `go generate .`。修改 YAML 后应重新生成并保证 `go build ./...` 通过。

## 命令行参数

以下为 `oapi-codegen -help` 所列出、日常最常用的选项（其余以 `--help` 为准）。

| 参数 | 说明 |
|------|------|
| `-config <file>` | YAML 配置文件路径；与下方「配置文件」一节对应 |
| `-generate <列表>` | 逗号分隔的生成目标；与配置里 `generate` 布尔项等价，见下表 |
| `-package <name>` | 生成代码的包名（配置文件中为 `package`） |
| `-o <path>` | 输出文件路径；未指定时输出到 stdout（配置文件中为 `output`） |
| `-import-mapping <dict>` | 外部 `$ref` 到 Go 包路径的映射（配置中为 `import-mapping`） |
| `-exclude-schemas` | 逗号分隔、不参与生成的 schema 名（配置中为 `output-options.exclude-schemas`） |
| `-exclude-tags` / `-include-tags` | 按 OpenAPI `tags` 过滤操作 |
| `-exclude-operation-ids` / `-include-operation-ids` | 按 `operationId` 过滤操作 |
| `-templates <dir>` | 自定义模板目录（配置中为 `output-options.user-templates`） |
| `-output-config` | 根据当前参数打印一份可用的 YAML 配置（便于初始化） |
| `-old-config-style` | 使用旧版配置格式（不推荐新工程使用） |
| `-version` | 打印版本并退出 |

### `-generate` 与配置键名的对应关系

命令行里写的是**短名**；YAML 里使用 **kebab-case** 布尔字段。**同一时间只能启用一种 Web 服务端生成器**（`chi-server`、`gin-server`、`echo-server` 等互斥）。

| `-generate` 取值 | 配置 `generate` 中的键 |
|------------------|-------------------------|
| `types` | 对应 **`models: true`**（生成 schema 模型） |
| `client` | `client` |
| `spec` | `embedded-spec` |
| `chi-server` | `chi-server` |
| `gin` | **`gin-server`** |
| `gorilla` | `gorilla-server` |
| `fiber` | `fiber-server` |
| `iris` | `iris-server` |
| `std-http` | `std-http-server` |
| `server` | Echo 服务端（详见上游 README；本仓库请用 `gin-server`） |
| `skip-fmt` | `output-options.skip-fmt` |
| `skip-prune` | `output-options.skip-prune` |

使用 **配置文件** 时，推荐只通过 `-config` 指定生成项，避免 CLI `-generate` 与 YAML 重复、难以排查。

## 配置文件（推荐格式）

配置文件与 `pkg/codegen.Configuration` 对齐，顶层字段包括：

| YAML 字段 | 说明 |
|-----------|------|
| `package` | 生成代码的包名（必填） |
| `output` | 输出 `.go` 文件路径（相对执行 `oapi-codegen` 时的当前工作目录） |
| `generate` | 见下一小节 |
| `compatibility` | 兼容旧版生成行为的开关（一般保持默认） |
| `output-options` | 过滤 schema/操作、模板、命名、Overlay 等 |
| `import-mapping` |  map：`外部引用路径` → `Go import 路径` |
| `additional-imports` | 额外 import 列表 |

### `generate` 块（布尔开关）

与本仓库相关的常用项：

| 键 | 含义 |
|----|------|
| `models` | 生成请求/响应等模型类型 |
| `embedded-spec` | 将规范嵌入代码（`GetSwagger` 等，便于文档 UI） |
| `gin-server` | 生成 Gin 的 `RegisterHandlers`、`ServerInterface` 等 |
| `chi-server` / `echo-server` / `fiber-server` / `iris-server` / `gorilla-server` / `std-http-server` | 其他框架的服务端骨架（与 `gin-server` **二选一**） |
| `client` | 生成客户端 |
| `strict-server` | 严格服务端包装（若需类型更严的 handler 签名） |
| `server-urls` | 为 `servers` 生成 URL 相关类型 |

### `output-options` 常用项

| 键 | 含义 |
|----|------|
| `include-tags` / `exclude-tags` | 字符串数组，按 tag 过滤 operation |
| `include-operation-ids` / `exclude-operation-ids` | 按 `operationId` 过滤 |
| `exclude-schemas` | 不生成指定 schema |
| `response-type-suffix` | 响应类型名后缀 |
| `skip-fmt` / `skip-prune` | 跳过 gofmt / 跳过剪枝优化 |
| `overlay` | OpenAPI Overlay（`path` + `strict`），在生成前改写规范 |

### `compatibility` 常用项

用于与旧版生成结果对齐（升级代码库时查阅上游 [README](https://github.com/oapi-codegen/oapi-codegen) 与各 issue）：如 `old-merge-schemas`、`old-enum-conflicts`、`old-aliasing` 等。

## 生成代码与运行时依赖

- 启用 **`embedded-spec`** 时，生成代码会依赖 **`github.com/getkin/kin-openapi/openapi3`**，需在模块中保留对应版本（`go mod tidy` 即可）。
- 启用 **`gin-server`** 时，会依赖 **`github.com/gin-gonic/gin`**。

## 快速生成一份配置草稿

在任意 OpenAPI 文件上，用当前 CLI 参数导出推荐 YAML：

```bash
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0 \
  -generate types,spec,gin -package apigen -output-config -o - api/openapi/openapi.yaml
```

将输出与现有 `api/openapi/oapi-codegen.yaml` 对比后合并即可。

## 与 Gin 应用接线

生成侧会提供 `ServerInterface` 与 `RegisterHandlers(router gin.IRouter, si ServerInterface)`。在应用中实现 `ServerInterface`（例如 `cmd/main.go` 里的 `openapiServer`），对根引擎或路由组调用 `RegisterHandlers` 即可。业务子路由（如 `/auth/oauth`）继续使用 `gin.RouterGroup` 挂载手写 handler，与生成路由并存。

## 如何打开 / 浏览 OpenAPI 文档

**规范本身（推荐）**是人读的 **`api/openapi/openapi.yaml`**：用任意编辑器打开即可；与生成代码是否重新跑过无关，它是单一事实来源。

**图形化浏览**可以任选其一：

- 在线 [Swagger Editor](https://editor.swagger.io)：菜单 **File → Import file**，选 `openapi.yaml`；或在编辑器里把文件内容粘贴进去。
- VS Code / Cursor 安装 OpenAPI / Swagger 类扩展（如 Redocly、42Crunch），在仓库里打开 `openapi.yaml` 即可预览。
- 本地起静态文档（需本机有 Node）：例如  
  `npx --yes @redocly/cli@latest preview-docs api/openapi/openapi.yaml`  
  会在本机起一个带 UI 的预览地址（终端会打印 URL）。

**Swagger / Try it out 里的 Servers**：规范里 `servers` 已包含 `http://127.0.0.1:11110`、`http://localhost:11110`、可改端口的 `http://127.0.0.1:{port}`，以及同源 `/`。若实际端口不同，在 UI 顶栏 **Servers** 里选带 `{port}` 的一项并修改，或直接在列表中选匹配项。改完 `openapi.yaml` 后需重新 `go generate ./api/gen/...` 才会更新嵌入 spec。

**生成代码里的「嵌入 spec」**（`apigen.GetSwagger()`）是给程序加载、校验或以后你自己挂 **`/openapi.json`** 等路由用的，**没有**自动在浏览器里打开；若希望「跑起服务就能点文档」，需要在 Gin 里增加路由（返回 YAML/JSON 或挂载 Swagger UI），这不是 oapi-codegen 默认行为。

## 延伸阅读

- 上游文档与更多示例：<https://github.com/oapi-codegen/oapi-codegen>
- 本仓库 API 分层与 OpenAPI 流程：`.cursor/rules/05-api.md`
