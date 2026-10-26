# Runtime & Observability

## 1. Môi trường chạy

| Môi trường | Cách chạy | Ghi chú |
|---|---|---|
| Local dev | `deployments/docker-compose.yml` (profiles `infra`, `full`, `observability`) | Postgres, Redis, RabbitMQ; 7 service; Nginx 8080 |
| Production | Một VPS, `deployments/docker-compose.prod.yml` | Image GHCR pin digest, migration job trước rollout, Nginx |

- `make docker-infra-up`: chỉ hạ tầng (Postgres/Redis/RabbitMQ).
- `make docker-dev-up`: full stack (infra + 7 service + Nginx).
- Chạy service trên host: `make run svc=<service>`; port HTTP 8101-8107, gRPC 9101-9104 (xem `architecture.md`).
- Compose nạp env vào container qua `env_file`: dev dùng `.env`, production dùng `.env.production`; key canonical nằm trong `.env.example` / `.env.production.example`.
- Certificate local nằm ngoài repo (`CERTS_DIR`), mount read-only `/run/littlecafe/certs`; chi tiết `docs/operations/certificates.md`.
- `APP_MODE=http|worker|all` chọn thành phần chạy trong một binary.

## 2. CI (`.github/workflows/ci.yml`)

Chạy khi PR/push `main` (bỏ qua `docs/**` và `*.md`):

| Job | Nội dung |
|---|---|
| static | `make fmt-check`, `make tidy-check`, `make generate`, kiểm tra generated committed |
| vet | `make vet`, `deployments/scripts/check-boundaries.sh` |
| test | `make test TEST_FLAGS=-short` |
| test-race | `make test-race TEST_FLAGS=-short` |
| contracts | `make buf-lint`, `make buf-gen` drift, `make openapi-gen`, Redocly lint OpenAPI, AsyncAPI validate, drift generated |
| image-build | build image `linux/amd64` (không push) |
| compose | `docker compose config --quiet` cho local + production compose |

## 3. CD (`.github/workflows/cd.yml`)

- Trigger: push tag `v*` hoặc `workflow_dispatch` với input `services` (`all` hoặc danh sách).
- Luồng: resolve plan → build/push GHCR (`:<sha>` + `:latest`, cache registry) → SSH upload compose/nginx → resolve digest immutable → cập nhật `/opt/littlecafe/.env.deploy` (`<SERVICE>_IMAGE=...@sha256:...`) → `compose pull` → chạy `migrate-<service>` (`--profile migrate`) → `up -d <service>` → recreate Nginx.
- Smoke: health container từng service + `https://api.littlecafe.kids/api/v1/health`.
- Rollback: dispatch lại với commit/tag cũ hoặc sửa digest trong `.env.deploy` rồi `up -d <service>`; chi tiết `docs/operations/deployment.md`.
- Secret (`DEPLOY_*`, `GHCR_*`) nằm trong GitHub Environment `production`; `/opt/littlecafe/.env.production` quản lý ngoài Git.

## 4. Logging

- `shared/pkg/logger`: Zap console (màu, dev) hoặc JSON (môi trường khác); file rotation qua lumberjack khi `logger.log_to_file`.
- `shared/observability/workerlog`: log chuẩn cho worker với `worker`, `triggerType`, `runId`, `durationMs`, counters và `result`.
- Error handler log Zap kèm `status`, `title`, `message`, `path`, `source` (stack frame nội bộ đã lọc), `cause` khi khác message.
- PII không được ghi log.

## 5. Metrics hiện có

`shared/infrastructure/messaging/metrics.go`:

- Counters: `publish_total`, `publish_failed`, `consume_ack_total`, `consume_nack_total`, `reconnect_total`.
- Gauges: `outbox_pending`, `queue_depth`, `dlq_depth` — cập nhật 30s bởi `OutboxWorker` qua `OutboxGaugeProbe` (đọc RabbitMQ management API).

**Trạng thái:** metrics chưa expose qua HTTP `/metrics`; hiện đọc qua `GetMetrics()` cho health/log. Không có Prometheus client instrumentation trong service.

## 6. OpenTelemetry & Prometheus (profile `observability`)

- `otel-collector` (contrib 0.138.0): nhận OTLP gRPC 4317/HTTP 4318, exporter `debug`.
- `prometheus` (v3.7.1): scrape chính collector (8888).
- **Trạng thái:** chỉ là scaffold; chưa có service nào gửi trace/metric, exporter mới ở mức debug. Bật profile này không ảnh hưởng runtime mặc định.

## 7. Vận hành

- Runbook triển khai/rollback: `docs/operations/deployment.md`.
- Certificate/mTLS: `docs/operations/certificates.md`.
- Setup local: `docs/development/environments.md` §4.
- Không dùng `docker compose down -v` trên production (xoá volume database); backup phải phủ cả 7 database trước thay đổi hạ tầng lớn.

## 8. Chưa triển khai

- App-level tracing/metrics qua OTel; endpoint `/metrics`; alerting.
- Health/readiness chi tiết theo dependency (hiện health trả `ok` tĩnh + gRPC health serving).
