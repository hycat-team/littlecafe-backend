# Service Boundaries & Contracts

## 1. Ranh giới service

Bảy service deploy độc lập; mỗi service sở hữu database riêng và chỉ service đó đọc/ghi database của mình.

| Service | Sở hữu nghiệp vụ | Database | gRPC host | Được phép gọi (read) |
|---|---|---|---|---|
| Identity | account, auth, Guardian-Child, RBAC | `littlecafe_identity` | có (9101) | - |
| Order | drink ordering, order lifecycle | `littlecafe_order` | có (9102) | Identity, Wallet, Activity |
| Activity | activity booking, competition, **Content** | `littlecafe_activity` | có (9103) | Identity, Order, Wallet |
| Wallet | ledger, top-up, payment, refund, withdrawal | `littlecafe_wallet` | có (9104) | Identity, Order, Activity |
| AI | RAG assistant, vector retrieval | `littlecafe_ai` | không | Identity, Activity (read-only), Order |
| Notification | gửi thông báo theo event | `littlecafe_notification` | không | - |
| Reporting | báo cáo, tổng hợp | `littlecafe_reporting` | không | read-only qua gRPC/event |

- Module `content` nằm trong Activity nhưng giữ ranh giới nghiệp vụ riêng; route `/api/v1/content/**` cũng trỏ về Activity.
- Danh sách "được phép gọi" là dự kiến theo workflow đã duyệt; từng call phải xuất hiện trong contract trước khi implement.

## 2. Quy tắc giao tiếp

- **Read-only xuyên service**: gRPC/protobuf trong `contracts/`. Không đọc bảng của service khác, kể cả cùng PostgreSQL instance.
- **Mutation xuyên service**: RabbitMQ event + transactional outbox (xem `event-driven-design.md`). Không gọi trực tiếp use case/service implementation của service khác.
- **Saga**: Order hoặc Activity sở hữu saga khởi tạo; Notification và Reporting tham gia theo choreography.
- **Error bridge**: `contracts/grpcx/errors.go` giữ status + message client-facing qua biên service; locale đi kèm qua metadata `x-littlecafe-locale`.
- **User identity**: khi hành động thay người dùng, JWT đi qua metadata `x-littlecafe-jwt` (`contracts/grpcx/metadata.go`).

## 3. Cấm import chéo

`deployments/scripts/check-boundaries.sh` (chạy trong CI) cưỡng chế:

1. `services/<a>` không import `services/<b>` (a ≠ b).
2. `shared` và `contracts` không import bất kỳ service module nào.
3. `domain` không import `application`, `infrastructure`, `presentation`.
4. `application` không import `infrastructure`, `presentation`.

Ngoài CI, layout `go.work` + `go.mod` per module khiến import chéo service trở thành lỗi build.

## 4. Contract map

| Loại | Vị trí | Trạng thái |
|---|---|---|
| HTTP | `contracts/openapi/components.yaml` + `generated.gen.go` (oapi-codegen strict server) | ProblemDetails + khung `/api/v1`; **chưa có endpoint nghiệp vụ** |
| Event | `contracts/asyncapi/asyncapi.yaml` | EventEnvelope + server; `channels`/`operations` rỗng |
| gRPC | `contracts/proto/` | chỉ có README; **chưa có proto file** |
| Helper gRPC | `contracts/grpcx/` (dial, metadata, mTLS, error bridge) | có |

Quy tắc:

- Contract là machine-readable source; implementation và generated code phải khớp. `generated.gen.go` được commit; CI regenerate và yêu cầu Git diff sạch.
- Không thêm route/RPC/event trước contract; không tạo contract suy đoán.
- Mọi thay đổi contract phải truy vết về requirement/workflow/ADR đã duyệt trong `littlecafe-docs`.

## 5. CI enforcement

- Job `static`: `make fmt-check`, `make tidy-check`, `make generate` + kiểm tra drift generated.
- Job `vet`: `make vet` + `check-boundaries.sh`.
- Job `contracts`: `make buf-lint`, `make buf-gen` drift, `make openapi-gen`, Redocly lint OpenAPI, AsyncAPI validate, kiểm tra generated committed.
- Job `build`/`image-build`: build từng service binary và image `linux/amd64`.

## 6. Health và discovery

- HTTP: `GET /api/v1/health` (Nginx route về identity).
- gRPC: health service chuẩn + reflection bật cho service host gRPC.
- Không service nào phụ thuộc trực tiếp endpoint của service khác ngoài contract.

## 7. Chưa triển khai

- Chưa có proto, RPC, event hay endpoint nghiệp vụ nào.
- Bảng "được phép gọi" chưa có call thực tế; bổ sung call phải đi kèm contract + test.
