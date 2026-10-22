# Môi trường phát triển và triển khai

Chuẩn cho developer và agent khi viết code, thêm config hoặc thay đổi hành vi
phụ thuộc môi trường. Runbook triển khai nằm trong
[`../operations/deployment.md`](../operations/deployment.md); branch/release nằm
trong [`git-workflow.md`](git-workflow.md).

## 1. Bảng môi trường

| Môi trường | Mục đích | Nhánh / trigger | URL | Thư mục VPS | Compose |
| --- | --- | --- | --- | --- | --- |
| Local | Phát triển, test nhanh | bất kỳ | `http://localhost:8080` | — | `deployments/docker-compose.yml` |
| Staging | Frontend test, UAT | push `develop` hoặc dispatch | `https://staging-api.littlecafe.kids:8443` | `/opt/littlecafe-staging` | `deployments/docker-compose.staging.yml` |
| Production | Người dùng thật | tag `v*` hoặc dispatch | `https://api.littlecafe.kids:8444` | `/opt/littlecafe` | `deployments/docker-compose.prod.yml` |

- Staging và production chạy trên cùng VPS, tách compose project, volume,
  network, secret, JWT và CA.
- Port 80/443 thuộc stack khác; littlecafe dùng 8081/8443 (staging) và
  8082/8444 (production).
- Local không cần mTLS; chỉ cần JWT nếu test đăng nhập.

## 2. Config và biến môi trường

### 2.1. Đặt giá trị ở đâu

| Loại giá trị | Nơi đặt | Ví dụ |
| --- | --- | --- |
| Hằng số ứng dụng | `configs/base.yaml` hoặc `configs/<service>.yaml` | port, timeout, rate limit, outbox batch size, provider name |
| Host/endpoint thay đổi theo môi trường | env hoặc Compose `environment` | `DB_HOST`, `REDIS_HOST`, `RABBITMQ_HOST` |
| Secret | env file trên máy/VPS hoặc file secret | DB password, provider API key, private key path |
| Định danh môi trường / flag vận hành | env | `APP_MODE`, `SERVER_ENV`, `LOG_LEVEL`, `EMAIL_ENABLED` |
| CORS frontend origins | `server.frontend_origins` (env `FRONTEND_ORIGINS`) | `https://a.example.com,https://b.example.com` |
| Chỉ dùng cho Compose | Compose hoặc env | `TZ`, mount path, `APP_UID`/`APP_GID` |

Nguyên tắc:

- Giá trị giống nhau ở mọi môi trường là literal trong `configs/`.
- Chỉ secret và giá trị thực sự khác theo môi trường mới đưa vào env.
- Không hardcode host, port, URL, credential trong code.
- Không rẽ nhánh business logic theo môi trường; dùng config hoặc feature flag.
- CORS xử lý ở Go middleware; `FRONTEND_ORIGINS` là CSV. `*` chỉ hợp lệ ngoài
  production; production phải liệt kê origin cụ thể.
- Không đọc `os.Getenv` trong `domain`, `application`, `presentation`; mọi giá
  trị đi qua struct config của `shared/config` hoặc config service. Validation
  fail fast khi khởi động.

### 2.2. Khi thêm hoặc đổi một setting

1. Xác định đúng loại theo bảng 2.1.
2. Cập nhật `configs/base.yaml` hoặc overlay service.
3. Cập nhật `.env.example`, `.env.production.example`, `.env.staging.example`.
4. Cập nhật Compose nếu setting ảnh hưởng `environment`, `ports`, `volumes`,
   `command` hoặc `user`.
5. Cập nhật `.github/workflows/ci.yml` nếu cần validate Compose mới.
6. Cập nhật `docs/operations/` và tài liệu này.
7. Thêm test cho config loader/validation nếu setting có ràng buộc.
8. Chạy `docker compose config` cho cả 3 môi trường.

## 3. Secret và file

- `.env`, `.env.production`, `.env.staging` không commit; chỉ nằm trên máy
  developer hoặc VPS, quyền `600`.
- File `.env*.example` chỉ chứa placeholder, không chứa giá trị thật.
- Comment trong file `.env*` viết bằng tiếng Anh.
- Vị trí secret trên VPS:
  - Production: `/opt/littlecafe-secrets/{certs,jwt,acme}` và `vercel.token`.
  - Staging: `/opt/littlecafe-staging-secrets/{certs,jwt}`.
  - TLS nginx: `/opt/littlecafe-tls/<domain>/{fullchain.pem,privkey.pem}`.
- `APP_UID`/`APP_GID` khớp user deploy (mặc định `1000`) để container đọc secret
  `700/600` mà không mở quyền toàn hệ thống.
- JWT ES256: mọi service mount public key; chỉ Identity mount private key.
- mTLS: 4 service gRPC (`identity/order/activity/wallet`) cần server cert;
  `service-client` dùng khi service gọi service khác qua gRPC.
- Token Vercel (DNS-01) và GHCR read token chỉ nằm trên VPS/CI secret.
- Không log secret; mask giá trị nhạy cảm khi debug.

## 4. Local setup

1. Copy `.env.example` thành `.env`, giữ file thật ngoài Git. Compose nạp `.env`
   vào mọi container qua `env_file`, nên biến provider (Google, payOS, Brevo,
   Cloudinary, JWT...) phải nằm trong file này.
2. Tạo certificate local ngoài repository bằng
   `deployments/scripts/rotate-mtls-certs.sh`; `CERTS_DIR` trong `.env` trỏ tới
   thư mục đó.
3. Chạy `docker compose -f deployments/docker-compose.yml --profile infra up -d`
   để bật PostgreSQL, Redis, RabbitMQ.
4. Chạy `make generate`, `make test`, `make build-all`.
5. Chạy full stack bằng profile `full` sau khi certificate sẵn sàng. Port host:
   HTTP 8101–8107, gRPC 9101–9104, Nginx 8080.
6. Chạy một service trên host bằng `make run svc=<service>`.

Không thêm entity/endpoint mẫu. Mỗi thay đổi nghiệp vụ bắt đầu từ requirement,
workflow, ADR và contract đã duyệt.

## 5. Database và migration

- Mỗi stack có Postgres riêng với 7 database/role: `littlecafe_<service>`.
- `database/init/00-init.sql` chỉ dùng cho local dev; staging/production dùng
  `deployments/scripts/init-db.sh` qua service `db-init` (idempotent).
- Migration Goose chạy qua service `migrate-<service>` với
  `command: ["-migrate-only"]`.
- Migration phải chạy trên database rỗng, idempotent và không đọc database của
  service khác.
- Không kết nối chéo môi trường (local → staging, staging → production).
- Thay đổi phá hủy schema phải có backup trước; ưu tiên forward-fix.
- Chi tiết Goose/schema: [`../system-design/data-and-migrations.md`](../system-design/data-and-migrations.md).

## 6. Provider bên thứ ba

| Provider | Local | Staging | Production |
| --- | --- | --- | --- |
| Google OAuth | trống | test client | client thật |
| payOS | trống | sandbox | thật |
| Brevo email | `EMAIL_ENABLED=false` | sandbox hoặc tắt | key thật |
| Cloudinary | trống | test account | thật |

- Không dùng credential production ở staging/local.
- Callback, return, cancel và webhook URL trỏ đúng môi trường.
- Staging không gửi email thật và không tạo giao dịch thật.

## 7. Kiểm tra theo môi trường

### Local

- Áp dụng [`gates.md`](gates.md) theo phạm vi thay đổi (gồm codegen/drift nếu
  đổi contract hoặc Wire).
- Kiểm tra Compose bằng `docker compose ... config` khi sửa file Compose.
- Chạy full stack khi thay đổi ảnh hưởng runtime; kiểm tra health và CORS.

### Staging

- Push `develop` để CD build, migrate, rollout, smoke.
- Kiểm tra container healthy, health qua nginx, CORS theo `FRONTEND_ORIGINS`,
  log không có lỗi cấu hình.
- Frontend test trên staging trước khi xem xét production.

### Production

- Merge `develop` → `main`, tag `v*`, theo dõi CD.
- Backup database trước thay đổi rủi ro.
- Smoke health, kiểm tra container, log và tài nguyên sau rollout.
- Rollback bằng digest/tag cũ; không `docker compose down -v`.

## 8. Pitfall đã gặp

- **Credential trong URL**: base64 có `/` hoặc `+` phải URL-encode khi build DSN
  (AMQP đã dùng `url.UserPassword`).
- **`-migrate-only`**: image có `ENTRYPOINT ["/app/app"]`, Compose `command` chỉ
  được là `["-migrate-only"]`.
- **SQL migration**: không để literal `\n` trong file `.sql`; psql meta-command
  là `\connect` (một backslash).
- **Quyền secret**: container distroless nonroot không đọc được thư mục `700`
  và key `600`; chạy service bằng `APP_UID/APP_GID` khớp user deploy.
- **Smoke từ GitHub runner**: firewall VPS có thể chặn IP runner ở 8443/8444;
  smoke public nên chạy từ VPS hoặc kiểm tra thủ công.
- **Nginx static upstream**: nếu service upstream chết, nginx có thể không start;
  kiểm tra container trước khi recreate nginx.

## 9. Checklist và liên kết

- Checklist/gate: [`gates.md`](gates.md). Trước merge vẫn kiểm tra thêm: 3 file
  `.env*.example`, Compose dev/staging/prod, CI Compose validation, secret, và
  migration trên DB rỗng.
- Deploy/rollback: [`../operations/deployment.md`](../operations/deployment.md);
  certificate: [`../operations/certificates.md`](../operations/certificates.md);
  branch/release: [`git-workflow.md`](git-workflow.md).
- Chính sách mức dự án: `.specify/memory/constitution.md` và
  `littlecafe-docs/deployment/environments.md`; runbook ngoài repo:
  `local-docs/deploy/`.
