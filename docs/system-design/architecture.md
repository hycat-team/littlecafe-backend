# Kiến trúc Backend LittleCafe

Tài liệu này là nguồn chuẩn về kiến trúc triển khai của backend. Quy tắc bắt buộc nằm trong `.specify/memory/constitution.md`; tài liệu này mô tả cấu trúc thực tế của code.

## 1. Workspace và topology

Backend là Go workspace (`go.work`, Go 1.26) gồm **9 module**: 7 service, `shared` và `contracts`. Module path chung là `github.com/hycat-team/littlecafe-backend`.

| Service | Module | Database | HTTP (local) | gRPC (local) |
|---|---|---|---|---|
| Identity | `services/identity` | `littlecafe_identity` | 8101 | 9101 |
| Order | `services/order` | `littlecafe_order` | 8102 | 9102 |
| Activity | `services/activity` | `littlecafe_activity` | 8103 | 9103 |
| Wallet | `services/wallet` | `littlecafe_wallet` | 8104 | 9104 |
| AI | `services/ai` | `littlecafe_ai` | 8105 | - |
| Notification | `services/notification` | `littlecafe_notification` | 8106 | - |
| Reporting | `services/reporting` | `littlecafe_reporting` | 8107 | - |

- Module `content` nằm trong Activity service nhưng giữ ranh giới nghiệp vụ riêng.
- Nginx local dùng port 8080; production dùng `https://api.littlecafe.kids/api/v1`.
- Mỗi service có Dockerfile riêng và deploy độc lập qua GHCR (xem `runtime-observability.md`).

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

`shared` chỉ chứa building block dùng chung, không thuộc service nào: `config`, `infrastructure` (db, messaging, caching, security, media, resilience, httptransport, audit), `application` (dto, event, sse, usecase, interface, constants), `presentation` (middleware, worker, response, sse), `domain` (entities, repositories, valueobject, constants), `observability`, `pkg`. Adapter của từng service nằm tại service sở hữu: Google OAuth ở Identity, Brevo ở Notification, payment provider ở Wallet, password/refresh-token crypto ở Identity, vector entity ở AI.

## 3. Phụ thuộc layer

Trong mỗi module: `presentation → application → domain ← infrastructure`.

- `domain` không import layer khác; chỉ entity, repository interface, constant.
- `application` chỉ phụ thuộc `domain`; không import `infrastructure`/`presentation`.
- `infrastructure` implement repository interface của `domain`.
- `presentation` gọi use case qua interface, không gọi repository trực tiếp.

`deployments/scripts/check-boundaries.sh` (chạy trong CI job `vet`) cưỡng chế:

1. Service không import package `services/<service khác>/`.
2. `shared` và `contracts` không import service.
3. `domain` không trỏ ra `application`/`infrastructure`/`presentation`.
4. `application` không trỏ vào `infrastructure`/`presentation`.

## 4. Bootstrap và vòng đời

- `internal/di/wire.go` khai báo `InitializeApp`; `wire_gen.go` được commit và CI kiểm tra drift (`make generate-check`).
- `internal/di/app.go` giữ các provider rỗng (`newModuleRouters`, `newModuleWorkers`, `newGRPCRegistrars`) cho tới khi có contract nghiệp vụ được duyệt.
- `bootstrap.NewRouter` dựng Gin engine, gắn `GlobalErrorHandler` và health probe `GET /api/v1/health`.
- `bootstrap.NewGRPCServer` bật health check và reflection, mount các registrar của module.
- `bootstrap.App` chạy HTTP + gRPC song song, start worker group, graceful shutdown theo SIGINT/SIGTERM (10s).
- `APP_MODE` chọn `http | worker | all`; mặc định `all`.

## 5. Cấu hình và hạ tầng

- `configs/base.yaml` là template chung; mỗi service merge thêm `configs/<service>.yaml` (section riêng của service), giá trị `${VAR:-default}` được thay bằng biến môi trường qua Viper. `shared/config/validation.go` kiểm tra bắt buộc (port, database, `JWT_ALGORITHM=ES256`, outbox batch size...).
- **PostgreSQL** (`pgvector/pgvector:pg17`): một instance, 7 database + 7 role tách biệt. `database/init/00-init.sql` chỉ tạo extension (`pgcrypto`, `vector` cho AI) khi volume khởi tạo lần đầu; không tạo bảng.
- **Redis**: cache, rate limit, cooldown, token blacklist, SSE Pub/Sub. Không phải nguồn dữ liệu authoritative.
- **RabbitMQ**: event bất đồng bộ theo outbox (xem `event-driven-design.md`).
- **Cloudinary**: media 2D. **Brevo**: email giao dịch. **payOS**: collection/payout (payout sau feature gate). **Google OAuth/OIDC**: đăng nhập.
- AI provider/model/vector store: xem `ai-service.md` — có mâu thuẫn nguồn cần chốt, agent không tự quyết.

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

Nginx sinh `X-Trace-ID` từ `$request_id` và forward cho mọi location; error handler đọc header này để điền `traceId` trong Problem Details.

## 7. Giao tiếp liên service

- **Read-only**: gRPC/protobuf qua `contracts/`; helper ở `contracts/grpcx` (dial, metadata, mTLS, error bridge).
- **Mutation**: RabbitMQ event + transactional outbox; không gọi trực tiếp database hay implementation của service khác.
- **Saga**: Order hoặc Activity sở hữu saga khởi tạo; Notification và Reporting tham gia choreography.
- Lỗi qua gRPC được bridge hai chiều bởi `contracts/grpcx/errors.go`: giữ status, message client-facing và `code` qua `google.rpc.ErrorInfo.Reason`; field errors dự trữ ở `google.rpc.BadRequest.FieldViolations` (details additive).

## 8. Chưa triển khai

Bootstrap hiện chỉ có health/readiness: không có entity, endpoint, RPC hay event nghiệp vụ. Các thư mục `domain/entities`, `application/usecase`, `presentation/handler|grpc|worker` của module đang rỗng. Bổ sung chỉ khi workflow, quyền sở hữu dữ liệu và contract tương ứng đã được duyệt (constitution I và III).
