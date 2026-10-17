# Event-Driven Design

## 1. Nguyên tắc

- Mutation liên service dùng RabbitMQ event; read-only dùng gRPC. Không service nào truy cập database của service khác.
- Mọi event publish qua **transactional outbox**: record ghi trong cùng transaction với thay đổi nghiệp vụ.
- Consumer xử lý **at-least-once** và dedup qua inbox.
- Event chỉ được đăng ký khi có contract AsyncAPI/JSON Schema xác định owner, version, payload và replay policy. Bootstrap hiện không có event nghiệp vụ nào.

## 2. Broker topology

- Exchange chuẩn: `littlecafe.events` (topic).
- Mỗi consumer sở hữu queue riêng; retry có giới hạn; dead-letter sang `<queue>.dlq`.
- Dead-letter exchange: `littlecafe.events.dlx` (topic). Mọi queue khai báo `x-dead-letter-exchange` + `x-dead-letter-routing-key = <queue>.dlq`.
- Queue quorum (tuỳ chọn) dùng `x-queue-type=quorum` với `x-delivery-limit=5`.
- `shared/infrastructure/messaging/topology.go` giữ registry `Topology` (rỗng cho tới khi có contract) và `DLQTopology` sinh tự động. Prefetch/concurrency mặc định lấy từ registry (fallback 20/1).
- Kết nối hỗ trợ TLS (`rabbitmq.tls_enabled`, `management_tls_enabled`); publish dùng publisher confirm.

## 3. Event envelope

Theo `contracts/asyncapi/asyncapi.yaml`, payload event gồm:

| Field | Kiểu | Ghi chú |
|---|---|---|
| `messageId` | uuid | dùng cho inbox dedup |
| `eventType` | string | tên event theo contract |
| `eventVersion` | integer ≥ 1 | version schema |
| `occurredAt` | date-time | thời điểm nghiệp vụ xảy ra |
| `producer` | string | service phát |
| `traceId` | string | nối với `X-Trace-ID` của request |
| `payload` | object | theo JSON Schema của event |

`correlationId` được inject vào context consumer khi nhận message (xem `RegisterTypedHandler`).

## 4. Outbox

Bảng `outbox_records` (mỗi service tự tạo trong migration bootstrap):

| Cột | Kiểu | Ghi chú |
|---|---|---|
| `id` | uuid | |
| `topic` | varchar(255) | routing key RabbitMQ |
| `payload` | jsonb | envelope |
| `headers` | jsonb | mặc định `{}` |
| `status` | varchar | `PENDING`, `PROCESSING`, `PUBLISHED`, `FAILED` |
| `attempts` | int | tăng khi mark failed |
| `last_error` | text | lỗi publish gần nhất |
| `created_at`, `updated_at`, `published_at` | timestamptz | |

Luồng publish:

1. Use case ghi outbox record trong transaction nghiệp vụ, sau đó `pg_notify('outbox_channel', '')`.
2. `OutboxWorker` nhận signal qua `OutboxListener` (`LISTEN outbox_channel`), hoặc poll fallback mỗi 2s.
3. `ClaimToPublish` claim batch bằng `FOR UPDATE SKIP LOCKED`, chuyển status sang `PROCESSING`; record `FAILED` quá backoff được claim lại; record `PROCESSING` kẹt quá 60s (worker crash) được reclaim.
4. Publish lên RabbitMQ; thành công → `PUBLISHED` + `published_at`; thất bại → `FAILED` + `attempts++` + `last_error`.
5. Cleanup định kỳ xoá record `PUBLISHED` cũ hơn `outbox.cleanup_retention_days` (mặc định 7 ngày).

Config (`outbox.*`): `retry_max_attempts` (10), `retry_interval_minutes` (1), `cleanup_retention_days` (7), `cleanup_interval_minutes` (60), `batch_size` (100), `listener_enabled` (true). LISTEN lỗi khi boot không làm fail service — chỉ mất độ trễ wake-up, không mất event.

## 5. Inbox và dedup

Bảng `inbox_messages` (`consumer`, `message_id`, `received_at`, `processed_at`) tạo trong migration bootstrap. Consumer ghi `messageId` vào inbox **trước khi** hoàn tất xử lý để đảm bảo at-least-once an toàn khi redelivery.

**Trạng thái:** chưa có consumer nghiệp vụ nào implement; đây là convention bắt buộc khi thêm consumer đầu tiên.

## 6. Consumer semantics

`RegisterTypedHandler(queue, handler, opts)` là entrypoint chuẩn:

- Resolve `Concurrency`/`Prefetch` từ `Topology`.
- Inject correlationId per-message vào context.
- Phân loại lỗi (khớp đúng 3 nhánh xử lý của `startConsumerLoop`):
  - `event.RetryableError` → requeue.
  - `event.PermanentError` → dead-letter.
  - Lỗi không phân loại → dead-letter kèm cảnh báo trong log; wrap bằng `event.RetryableError` nếu redelivery có thể giúp.

## 7. Metrics

`shared/infrastructure/messaging/metrics.go` cung cấp counter/gauge để `/metrics` và health probe đọc:

- Counters: `publish_total`, `publish_failed`, `consume_ack_total`, `consume_nack_total`, `reconnect_total`.
- Gauges (refresh 30s bởi outbox worker): `outbox_pending`, `queue_depth`, `dlq_depth`.

## 8. Realtime (khác RabbitMQ)

SSE dùng Redis Pub/Sub làm relay giữa các instance (`redis_realtime_publisher.go` → `SSERelay`), không dùng RabbitMQ. Redis chỉ là kênh truyền; trạng thái chuẩn vẫn nằm ở DB và REST fallback.

## 9. Chưa triển khai

- `Topology` rỗng: chưa có queue/event nào được duyệt.
- Chưa có saga worker, consumer nghiệp vụ, hay replay policy cụ thể.
- Thêm event đầu tiên cần: AsyncAPI/JSON Schema + owner + version + payload + replay policy, migration bảng nghiệp vụ, handler phân loại `RetryableError`/`PermanentError`.
