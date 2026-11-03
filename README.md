# LittleCafe Backend

Contract-first backend for LittleCafe, a kids cafe platform. Built as a Go
multi-module workspace (`go.work`) with **four independently deployable
services**, Clean Architecture, and compiler-enforced service boundaries.
Module path: `github.com/hycat-team/littlecafe-backend`.

## Workspace

- Four services: Commerce, Experience, AI, Platform.
- `Commerce` owns Identity, Order, and Wallet modules; `Experience` owns
  Activity and Content; `Platform` owns Notification and Reporting.
- `shared/` holds common infrastructure; `contracts/` is the machine-readable
  source for OpenAPI, AsyncAPI/JSON Schema, and protobuf.
- One PostgreSQL instance with four databases and four separate roles.

| Service | HTTP (local) | gRPC (local) |
|---|---:|---:|
| Commerce | 8101 | 9101 |
| Experience | 8102 | 9102 |
| AI | 8103 | 9103 |
| Platform | 8104 | 9104 |

Nginx local listens on 8080. Production API:
`https://api.littlecafe.kids:8444/api/v1`; staging:
`https://staging-api.littlecafe.kids:8443/api/v1`.

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.26 multi-module workspace (`go.work`) |
| HTTP | Gin + `oapi-codegen` strict server |
| Contracts | OpenAPI, AsyncAPI/JSON Schema, protobuf |
| Inter-service | gRPC for synchronous reads/commands; RabbitMQ + transactional outbox for asynchronous events |
| Database | PostgreSQL 17 + pgvector (4 databases/roles) |
| Cache / realtime | Redis 7 |
| Migration / DI | Goose / Google Wire |
| Providers | Google OAuth, payOS, Brevo, Cloudinary |
| Observability | Zap, Prometheus/OTel scaffolding |
| CI/CD | GitHub Actions; Docker Compose + Nginx on VPS |

## Prerequisites

- Go `>= 1.26`
- Docker & Docker Compose
- GNU Make (Windows: Chocolatey/Git Bash/WSL)
- Buf CLI (`make install-tools`) when changing protobuf
- OpenSSL (local JWT keys)

## Quickstart

```powershell
make env-local
make docker-infra-up
make generate && make test && make build-all
```

Checklist đầy đủ (JWT, full stack, mTLS, quyền đọc env):
**[docs/development/bootstrap.md](docs/development/bootstrap.md)**.
Chi tiết config/port: [docs/development/environments.md](docs/development/environments.md).
Chạy `make help` để xem toàn bộ target.

## Docs

- [Bootstrap](docs/development/bootstrap.md)
- [Documentation index](docs/README.md)
- [Spec Kit workflow](docs/development/spec-kit-workflow.md)
- [Git workflow](docs/development/git-workflow.md)
- [Gates](docs/development/gates.md)
- [Environments](docs/development/environments.md)
- [SOPS keys](docs/development/sops-keys.md)
- [Architecture](docs/system-design/architecture.md)
- [Design patterns](docs/system-design/design-pattern.md)
- [Deployment](docs/operations/deployment.md)

## Notes

- Contracts live in `contracts/`; generated code is committed and CI verifies
  drift.
- Errors use RFC 9457 Problem Details with `code`, `traceId`, and field errors.
- Client-facing messages are localized via `Accept-Language: vi|en` (default
  `vi`).
- There are no sample business APIs; bootstrap only exposes health/readiness.
- AI provider/model/vector store decisions are not fully settled; read
  [docs/system-design/ai-service.md](docs/system-design/ai-service.md) before
  implementing AI-related work.
