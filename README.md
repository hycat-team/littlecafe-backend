# LittleCafe Backend

Contract-first backend for LittleCafe, a kids cafe platform. Built as a Go
multi-module workspace (`go.work`) with **seven independently deployable
services**, Clean Architecture, and compiler-enforced service boundaries.
Module path: `github.com/hycat-team/littlecafe-backend`.

---

## Workspace

- Seven services: Identity, Order, Activity, Wallet, AI, Notification, Reporting.
- `Content` is a module inside the Activity service with its own business boundary.
- `shared/` holds common infrastructure; `contracts/` is the machine-readable
  source for OpenAPI, AsyncAPI/JSON Schema, and protobuf.
- One PostgreSQL instance with seven databases and seven separate roles.

| Service | HTTP (local) | gRPC (local) |
|---|---:|---:|
| Identity | 8101 | 9101 |
| Order | 8102 | 9102 |
| Activity | 8103 | 9103 |
| Wallet | 8104 | 9104 |
| AI | 8105 | - |
| Notification | 8106 | - |
| Reporting | 8107 | - |

Nginx local listens on 8080. Production API is
`https://api.littlecafe.kids/api/v1`; the public website is
`https://littlecafe.kids`.

---

## Tech Stack

| Component | Technology |
|---|---|
| Language | **Go 1.26** multi-module workspace (`go.work`, 9 modules) |
| HTTP | **Gin** + **oapi-codegen** strict server interface |
| Contracts | **OpenAPI** (HTTP), **AsyncAPI/JSON Schema** (events), **protobuf** (read-only cross-service) |
| Inter-service | **gRPC** for reads; **RabbitMQ** + transactional outbox for mutations |
| Database | **PostgreSQL 17 + pgvector** (7 databases, 7 roles) |
| Cache / realtime | **Redis 7** (cache, rate limit, cooldown, token blacklist, SSE Pub/Sub) |
| Migration | **Goose** per service, embedded SQL |
| DI | **Google Wire** per service |
| Media / payments / email / auth | **Cloudinary** / **payOS** / **Brevo** / **Google OAuth** |
| Logging & metrics | **Zap** JSON, worker run logs, Prometheus/OTel scaffolding |
| CI/CD | **GitHub Actions** (CI matrix; CD to VPS via GHCR pinned by digest) |
| Deploy | **Docker Compose** on one VPS + **Nginx** |

---

## Prerequisites

- **Go** `>= 1.26`
- **Docker & Docker Compose** — Postgres, Redis, RabbitMQ locally
- **GNU Make** — built-in on macOS/Linux; on Windows use Chocolatey/Git Bash/WSL
- **Buf CLI** (`make install-tools`) — only needed when changing protobuf contracts
- Certificates for the full local stack (see below)

---

## Local Setup

```powershell
Copy-Item .env.example .env
docker compose -f deployments/docker-compose.yml --profile infra up -d
make generate
make test
make build-all
```

- `make docker-infra-up` starts only PostgreSQL, Redis, and RabbitMQ.
- `make docker-dev-up` starts the full stack (infra + services + Nginx).
- The full stack requires local certificates at `CERTS_DIR` outside the
  repository, created by `deployments/scripts/rotate-mtls-certs.sh`, then run the
  `full` profile.
- Run one service on the host with `make run svc=<service>`.
- There are no sample business APIs; the bootstrap only exposes health/readiness
  and infrastructure wiring.

---

## Makefile Targets

```bash
make wire svc= / wire-all          # Google Wire generation
make buf-gen / buf-lint            # protobuf contracts
make openapi-gen                   # oapi-codegen strict server types
make generate                      # wire-all + openapi-gen
make generate-check                # regenerate and verify no drift
make fmt-check / tidy-check / vet  # static checks
make test / test-race              # workspace tests
make lint                          # fmt-check + vet
make check                         # full gate (fmt, tidy, drift, vet, race, contract lint, build)
make build svc= / build-all        # build one / all services
make run svc=                      # run one service locally
make dev                           # generate + build all
make migrate-create svc= name=     # Goose per service
make migrate-up|status|down|version svc=
make docker-infra-up|docker-dev-up|docker-dev-stop|docker-dev-down
make nginx-reload-local
```

---

## API Documentation

- Contracts live in `contracts/`; `contracts/openapi/generated.gen.go` is
  committed and CI verifies drift.
- Errors are RFC 9457 Problem Details (`application/problem+json`) with `code`,
  `traceId`, and field errors.
- Important mutations require an `Idempotency-Key`.
- See [contract conventions](docs/api/contracts.md).

---

## Project Structure

```text
.
├── go.work                     # 9 modules: 7 services + shared + contracts
├── services/<name>/            # cmd, internal/{bootstrap,di,modules}, database/migrations, Dockerfile
├── shared/                     # config, infrastructure, application, presentation, domain, observability, pkg
├── contracts/                  # OpenAPI, AsyncAPI, proto, grpcx helpers
├── deployments/                # compose (dev/prod), nginx, observability, scripts
├── database/init/              # dev-only role/database/extension bootstrap (no tables)
├── configs/                    # base.yaml + per-service overlay (env-substituted)
├── docs/                       # system-design, api, development, operations
└── .github/workflows/          # ci.yml, cd.yml
```

Each module follows Clean Architecture
(`domain` ← `application` ← `presentation`, `infrastructure` implements
`domain`). Cross-service calls go through `contracts/` only. See
[docs/system-design/](docs/README.md) and the constitution
(`.specify/memory/constitution.md`).

---

## Deployment & Operations

- CD workflow `.github/workflows/cd.yml` runs on `v*` tags or manual dispatch,
  pushes images to GHCR, pins digests in `/opt/littlecafe/.env.deploy`, runs the
  per-service migration job, and rolls out services behind Nginx.
- Rollback: dispatch with the previous commit/tag or restore the image digest and
  re-run `up -d <service>`.
- Runbooks: [deployment](docs/operations/deployment.md),
  [certificates](docs/operations/certificates.md).

---

## Further Reading

- [Documentation index](docs/README.md)
- [System design](docs/system-design/architecture.md) — topology, boundaries, events, security
- [Design patterns](docs/system-design/design-pattern.md) — error shape, handler, repository, worker
- [Agent workflow](docs/development/agent-workflow.md) and [AGENTS.md](AGENTS.md)

AI provider/model/vector store decisions are not fully settled; see
[docs/system-design/ai-service.md](docs/system-design/ai-service.md) before
implementing anything AI-related.
