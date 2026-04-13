# blink
A lightweight, high-performance microblogging platform designed for sharing fleeting thoughts, daily moments, and visual stories.

## Run locally

Requires Redis (`127.0.0.1:6379` by default) and Go. From the repository root:

```bash
mkdir -p data
go run ./cmd
```

See [docs/run-local.md](docs/run-local.md) for environment variables, health checks, migrations, and IDE launch notes.

Architecture (layers, HTTP vs notification flow): [docs/architecture.md](docs/architecture.md).
