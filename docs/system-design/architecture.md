# Kiến trúc Backend LittleCafe

Nguồn chuẩn về kiến trúc triển khai của backend. Quy tắc bắt buộc nằm trong
`.specify/memory/constitution.md`; tài liệu này mô tả cấu trúc thực tế của code,
ranh giới service và contract.

## 1. Workspace và topology

Backend là Go workspace (`go.work`, Go 1.26) gồm **9 module**: 7 service,
`shared` và `contracts`. Module path chung:
`github.com/hycat-team/littlecafe-backend`.

| Service | Module | Database | HTTP (local) | gRPC (local) | Được phép gọi (read) |
|---|---|---|---|---|---|
| Identity | `services/identity` | `littlecafe_identity` | 8101 | 9101 | - |
| Order | `services/order` | `littlecafe_order` | 8102 | 9102 | Identity, Wallet, Activity |
| Activity | `services/activity` | `littlecafe_activity` | 8103 | 9103 | Identity, Order, Wallet |
| Wallet | `services/wallet` | `littlecafe_wallet` | 8104 | 9104 | Identity, Order, Activity |
| AI | `services/ai` | `littlecafe_ai` | 8105 | - | Identity, Activity (read-only), Order |
| Notification | `services/notification` | `littlecafe_notification` | 8106 | - | - |
| Reporting | `services/reporting` | `littlecafe_reporting` | 8107 | - | read-only qua gRPC/event |

- Module `content` nằm trong Activity nhưng giữ ranh giới nghiệp vụ riêng; route
  `/api/v1/content/**` trỏ về Activity.
- Nginx local dùng port 8080; production dùng
  `https://api.littlecafe.kids/api/v1`.
- Mỗi service có Dockerfile riêng và deploy độc lập qua GHCR.
- Danh sách "được phép gọi" là dự kiến theo workflow đã duyệt; từng call phải
  xuất hiện trong contract trước khi implement.

## 2. Layout module chuẩn

```text
services/<service>/
  cmd/<service>/main.go          # entrypoint
  internal/bootstrap/            # router, gRPC server, worker group, app lifecycle
  internal/di/                   # Wire providers (wire.go, app.go, wire_gen.go)
  internal/modules/<module>/
    domain/{entities,repositories}
    application/{ports,usecase}
    infrastructure/persistence
    presentation/{handler,router,grpc,worker}
  database/migrations/           # Goose SQL + embed FS
  api/swagger/                   # output swagger (khi có)
```

`shared` chỉ chứa building block dùng chung, không thuộc service nào: `config`,
`infrastructure` (db, messaging, caching, security, media, resilience,
httptransport, audit), `application` (dto, event, sse, usecase, interface,
constants), `presentation` (middleware, worker, response, sse), `domain`
(entities, repositories, valueobject, constants), `observability`, `pkg`.
Adapter của từng service nằm tại service sở hữu: Google OAuth ở Identity, Brevo
ở Notification, payment provider ở Wallet, password/refresh-token crypto ở
Identity, vector entity ở AI.

## 3. Phụ thuộc layer

Trong mỗi module: `presentation → application → domain ← infrastructure`.

- `domain` không import layer khác; chỉ entity, repository interface, constant.
- `application` chỉ phụ thuộc `domain`; không import `infrastructure`/`presentation`.
- `infrastructure` implement repository interface của `domain`.
- `presentation` gọi use case qua interface, không gọi repository trực tiếp.

`deployments/scripts/check-boundaries.sh` (CI job `vet`) cưỡng chế:

1. `services/<a>` không import `services/<b>` (a ≠ b).
2. `shared` và `contracts` không import service.
3. `domain` không import `application`, `infrastructure`, `presentation`.
4. `application` không import `infrastructure`, `presentation`.

Ngoài CI, layout `go.work` + `go.mod` per module khiến import chéo service trở
thành lỗi build.

## 4. Bootstrap và vòng đời

- `internal/di/wire.go` khai báo `InitializeApp`; `wire_gen.go` được commit và
  CI kiểm tra drift (`make generate-check`).
- `internal/di/app.go` giữ provider rỗng (`newModuleRouters`,
  `newModuleWorkers`, `newGRPCRegistrars`) cho tới khi có contract nghiệp vụ.
- `bootstrap.NewRouter` dựng Gin engine, gắn `GlobalErrorHandler` và health probe
  `GET /api/v1/health`.
- `bootstrap.NewGRPCServer` bật health check và reflection, mount registrar.
- `bootstrap.App` chạy HTTP + gRPC song song, start worker group, graceful
  shutdown theo SIGINT/SIGTERM (10s).
- `APP_MODE` chọn `http | worker | all`; mặc định `all`.

## 5. Cấu hình và hạ tầng

- `configs/base.yaml` là template chung; mỗi service merge thêm
  `configs/<service>.yaml`. Hằng số nằm literal trong config; chỉ secret và giá
  trị đặc thù môi trường mới dùng `${VAR}` qua Viper. `shared/config/validation.go`
  kiểm tra bắt buộc (port, database, `jwt.algorithm=ES256`, outbox batch size...).
- **PostgreSQL** (`pgvector/pgvector:pg17`): một instance, 7 database + 7 role
  tách biệt. `database/init/00-init.sql` chỉ tạo extension (`pgcrypto`, `vector`
  cho AI) khi volume khởi tạo lần đầu; không tạo bảng.
- **Redis**: cache, rate limit, cooldown, token blacklist, SSE Pub/Sub; không là
  nguồn dữ liệu authoritative.
- **RabbitMQ**: event bất đồng bộ theo outbox
  ([`event-driven-design.md`](event-driven-design.md)).
- **Cloudinary**: media 2D. **Brevo**: email giao dịch. **payOS**: collection/
  payout. **Google OAuth/OIDC**: đăng nhập.
- AI provider/model/vector store: xem [`ai-service.md`](ai-service.md) — có mâu
  thuẫn nguồn cần chốt, agent không tự quyết.

## 6. Routing

Nginx (`deployments/nginx/local.conf`, `prod.conf`) route theo prefix:

| Path | Service |
|---|---|
| `/api/v1/health` | identity |
| `/api/v1/identity/**` | identity |
| `/api/v1/order/**` | order |
| `/api/v1/activity/**`, `/api/v1/content/**` | activity |
| `/api/v1/wallet/**` | wallet |
| `/api/v1/ai/**` | ai |
| `/api/v1/notification/**` | notification |
| `/api/v1/reporting/**` | reporting |

Nginx sinh `X-Trace-ID` từ `$request_id` và forward cho mọi location; error
handler đọc header này để điền `traceId` trong Problem Details.

## 7. Ranh giới và giao tiếp liên service

- **Read-only xuyên service**: gRPC/protobuf trong `contracts/`; helper ở
  `contracts/grpcx` (dial, metadata, mTLS, error bridge). Không đọc bảng của
  service khác, kể cả cùng PostgreSQL instance.
- **Mutation xuyên service**: RabbitMQ event + transactional outbox; không gọi
  trực tiếp use case/service implementation của service khác.
- **Saga**: Order hoặc Activity sở hữu saga khởi tạo; Notification và Reporting
  tham gia theo choreography.
- **Error bridge**: `contracts/grpcx/errors.go` giữ status, message client-facing
  và `code` qua `google.rpc.ErrorInfo.Reason`; locale lan qua metadata
  `x-littlecafe-locale` và `ErrorInfo.Metadata` (`messageKey`, `locale`); field
  errors dự trữ ở `google.rpc.BadRequest.FieldViolations` (details additive).
- **User identity**: khi hành động thay người dùng, JWT đi qua metadata
  `x-littlecafe-jwt` (`contracts/grpcx/metadata.go`). HTTP ưu tiên Bearer, fallback
  cookie `access_token`; refresh token dùng cookie `refresh_token`.
- **Idempotency**: Redis store dùng chung, scope user+method+path, TTL 24h;
  middleware fail closed (`503`) khi Redis lỗi.

## 8. Contract map

| Loại | Vị trí | Trạng thái |
|---|---|---|
| HTTP | `contracts/openapi/components.yaml` + `generated.gen.go` (oapi-codegen strict server) | ProblemDetails, SuccessResponse, CursorPage, AdminPage, IdempotencyKey, securitySchemes; khung `/api/v1`; chưa có endpoint nghiệp vụ |
| Event | `contracts/asyncapi/asyncapi.yaml` | EventEnvelope + server; `channels`/`operations` rỗng |
| gRPC | `contracts/proto/` | chỉ có README; chưa có proto file |
| Helper gRPC | `contracts/grpcx/` | có |

Quy tắc:

- Contract là machine-readable source; implementation và generated code phải
  khớp. `generated.gen.go` được commit; CI regenerate và yêu cầu Git diff sạch.
- Không thêm route/RPC/event trước contract; không tạo contract suy đoán.
- Mọi thay đổi contract phải truy vết về requirement/workflow/ADR đã duyệt
  trong `littlecafe-docs`.

## 9. CI enforcement

Gate và CI command: [`../development/gates.md`](../development/gates.md) và
`.github/workflows/ci.yml`. Boundary/import check chạy trong job `vet` bằng
`deployments/scripts/check-boundaries.sh`. Contract lint chạy `make contract-lint`
(buf + Redocly + AsyncAPI + error-code registry) và job `contracts` kiểm tra
generated drift.

## 10. Health và discovery

- HTTP: `GET /api/v1/health` (Nginx route về identity).
- gRPC: health service chuẩn + reflection bật cho service host gRPC.
- Không service nào phụ thuộc trực tiếp endpoint của service khác ngoài contract.

## 11. Chưa triển khai

- Bootstrap chỉ có health/readiness: chưa có entity, endpoint, RPC hay event
  nghiệp vụ; các thư mục `domain/entities`, `application/usecase`,
  `presentation/handler|grpc|worker` đang rỗng.
- Bảng "được phép gọi" chưa có call thực tế; bổ sung call phải đi kèm contract +
  test, sau khi workflow/quyền sở hữu dữ liệu được duyệt (constitution I và III).
