# Data & Migrations

## 1. Database topology

- Một PostgreSQL instance (`pgvector/pgvector:pg17`) chạy 7 database tách biệt, mỗi database có role riêng trùng tên service.
- `database/init/00-init.sql` chạy **một lần** khi volume khởi tạo: tạo role, database và extension (`pgcrypto` cho mọi database, `vector` cho AI). File này **không tạo bảng** và là bootstrap dev; production provision role/database/extension một lần theo `docs/operations/deployment.md` mục "Provision database lần đầu", credential quản lý ngoài Git.
- Kết nối app dùng `database.*` trong `configs/config.yaml`; `TimeZone=Asia/Ho_Chi_Minh`; `sslmode` cấu hình theo môi trường.

## 2. Quyền sở hữu dữ liệu

- Mỗi service chỉ đọc/ghi database của mình; không truy vấn chéo, không join chéo, không import entity/repository của service khác.
- Dữ liệu liên service trao đổi qua gRPC (read) hoặc event (mutation).
- Bảng `outbox_records`, `inbox_messages`, `audit_log` được mỗi service tạo riêng trong migration bootstrap của service đó.

## 3. Migrations (Goose)

- Mỗi service có `services/<service>/database/migrations/` với `migrations.go` (embed `*.sql`) và file tuần tự `00001_bootstrap.sql`, `00002_...`.
- File migration theo format Goose: `-- +goose Up` / `-- +goose Down`.
- `bootstrap` hiện tạo: `outbox_records` (+ index `idx_outbox_pending`), `inbox_messages` (PK `consumer, message_id`), `audit_log` (+ index `idx_audit_log_occurred_at`).
- Chạy migration:
  - Local/dev: `DB_MIGRATE_ON_START=true` (compose dev bật) → `shared/infrastructure/db/migrations.go` chạy `goose.Up` khi service khởi động, dùng **direct DSN** (không qua pooler), `pg_advisory_lock(hashtext('littlecafe:migrate:<db>'))`, tối đa 1 connection.
  - Production: CD chạy job `migrate-<service>` (`--profile migrate`) trước khi rollout service tương ứng.
  - CLI: `make migrate-create svc=<service> name=<change>`, `make migrate-up svc=`, `make migrate-status svc=`, `make migrate-down svc=`, `make migrate-version svc=`.
- **Không sửa migration đã áp dụng ở môi trường dùng chung/live.** Thay đổi schema mới là file mới.
- Không đặt DDL nghiệp vụ trong `database/init`.

## 4. Quy ước dữ liệu

| Chủ đề | Quy ước | Nguồn code |
|---|---|---|
| ID | UUIDv7 | `shared/pkg/id` (`uuid.NewV7`) |
| Entity base | `BaseEntity{id uuid, created_at timestamptz}`; `AuditableEntity` thêm `updated_at`; `SoftDeleteEntity` thêm `is_deleted` + `SoftDelete()/Undo()` | `shared/domain/entities/base.go` |
| Tiền | `MoneyVND int64`, không âm; cấm floating-point | `shared/domain/valueobject/money.go` |
| Thời gian | `TIMESTAMPTZ`; business calendar `Asia/Ho_Chi_Minh`; API RFC3339 offset +07:00 | constitution IV |
| Vector | `entities.Vector []float32` map sang cột `vector` (pgvector) | `shared/domain/entities/vector.go` |
| JSONB | `JSONField` trả `string` khi ghi (tương thích pooler) | `shared/domain/repositories/outbox.go` |

- GORM tag dùng `timestamptz` cho cột thời gian; ví dụ `outbox_records`.
- Không tạo bảng/entity giả để minh họa; mỗi bảng phải thuộc workflow + contract đã duyệt.

## 5. Transaction và consistency

- Transaction qua `IUnitOfWork.Execute`; transaction truyền bằng `context.Context` (`db.InjectTx`/`GetTx`), mọi query dùng `WithContext`.
- Mutation quan trọng dùng `Idempotency-Key` (API) và outbox/inbox cho event.
- Unique constraint/optimistic locking là trách nhiệm của module khi có nghiệp vụ thật.

## 6. Chưa triển khai

- Chưa có bảng nghiệp vụ nào; chỉ có 3 bảng hạ tầng ở §3.
- Chưa có migration cho ledger, booking, order, guardian link... — bổ sung cùng feature tương ứng.
- Chưa có công cụ kiểm tra ownership DDL/migration chéo service; hiện cưỡng chế bằng review + `check-boundaries.sh` cho import.
