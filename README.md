# blink
A lightweight, high-performance microblogging platform designed for sharing fleeting thoughts, daily moments, and visual stories.

## Run locally

Requires Redis (`127.0.0.1:6379` by default) and Go. From the repository root:

```bash
mkdir -p data
go run ./cmd
```

See [docs/run-local.md](docs/run-local.md) for environment variables, health checks, migrations, Git worktree usage, and IDE launch notes.

Architecture (layers, HTTP vs notification flow): [docs/architecture.md](docs/architecture.md).

HTTP 契约与 OpenAPI：规范在 [`api/openapi/openapi.yaml`](api/openapi/openapi.yaml)；修改后按 [docs/oapi-codegen.md](docs/oapi-codegen.md) 中的命令重新生成 `api/gen/apigen.gen.go`。
后台用 golang，后台管理用 html写了一些页面，前端打算用 flutter 进行编写
后端用了ddd的开发模式，数据库可以根据自己的需要切换为postgresql或者sqlite,用gromm来做ORM。
后台管理页面用的最原始的html,没有使用react等框架，比较简单干净
