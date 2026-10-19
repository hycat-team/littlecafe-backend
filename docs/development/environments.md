# Môi trường phát triển và triển khai

Tài liệu này là chuẩn cho developer và agent khi viết code, thêm config hoặc thay
đổi hành vi phụ thuộc môi trường. Quy trình vận hành chi tiết nằm trong
`docs/operations/` và runbook `local-docs/deploy/`; tài liệu này chỉ nêu quy tắc
để code chạy đúng ở **local**, **staging** và **production**.

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
| Hằng số của ứng dụng | `configs/base.yaml` hoặc `configs/<service>.yaml` | port, timeout, rate limit, outbox batch size, provider name |
| Host/endpoint thay đổi theo môi trường | env hoặc Compose `environment` | `DB_HOST`, `REDIS_HOST`, `RABBITMQ_HOST` |
| Secret | env file trên máy/VPS hoặc file secret | DB password, provider API key, private key path |
| Định danh môi trường / flag vận hành | env | `APP_MODE`, `SERVER_ENV`, `LOG_LEVEL`, `EMAIL_ENABLED` |
| Chỉ dùng cho Compose | Compose hoặc env | `TZ`, mount path, `APP_UID`/`APP_GID` |

Nguyên tắc:

- Giá trị **giống nhau ở mọi môi trường** phải là literal trong `configs/`.
- Chỉ secret và giá trị **thực sự khác theo môi trường** mới đưa vào env.
- Không hardcode host, port, URL, credential trong code.
- Không rẽ nhánh business logic theo môi trường; dùng config hoặc feature flag.

### 2.2. Khi thêm hoặc đổi một setting

1. Xác định đúng loại theo bảng 2.1.
2. Cập nhật `configs/base.yaml` hoặc overlay service.
3. Cập nhật cả 3 file mẫu: `.env.example`, `.env.production.example`,
   `.env.staging.example`.
4. Cập nhật Compose nếu setting ảnh hưởng `environment`, `ports`, `volumes`,
   `command` hoặc `user`: `docker-compose.yml`, `docker-compose.prod.yml`,
   `docker-compose.staging.yml`.
5. Cập nhật `.github/workflows/ci.yml` nếu cần validate Compose mới.
6. Cập nhật `docs/operations/` và tài liệu này.
7. Thêm test cho config loader hoặc validation nếu setting có ràng buộc.
8. Chạy `docker compose config` cho cả 3 môi trường.

Không đọc `os.Getenv` trong `domain`, `application` hay `presentation`; mọi giá
trị đi qua struct config của `shared/config` hoặc config service. Validation phải
fail fast khi khởi động.

## 3. Secret và file

- `.env`, `.env.production`, `.env.staging` **không commit**; chỉ nằm trên máy
  developer hoặc VPS, quyền `600`.
- File `.env*.example` trong repo chỉ chứa placeholder, không chứa giá trị thật.
- Comment trong file `.env*` viết bằng tiếng Anh.
- Vị trí secret trên VPS:
  - Production: `/opt/littlecafe-secrets/{certs,jwt,acme}` và `vercel.token`.
  - Staging: `/opt/littlecafe-staging-secrets/{certs,jwt}`.
  - TLS nginx: `/opt/littlecafe-tls/<domain>/{fullchain.pem,privkey.pem}`.
- `APP_UID`/`APP_GID` phải khớp user deploy (mặc định `1000`) để container đọc
  được secret `700/600` mà không phải mở quyền cho toàn hệ thống.
- JWT ES256: mọi service mount public key; chỉ Identity mount private key.
- mTLS: 4 service gRPC (`identity/order/activity/wallet`) cần server cert;
  `service-client` dùng khi service gọi service khác qua gRPC.
- Token Vercel (DNS-01) và GHCR read token chỉ nằm trên VPS/CI secret, không nằm
  trong repo hay env của service.
- Không log secret. Nếu cần debug, mask toàn bộ giá trị nhạy cảm.

## 4. Database và migration

- Mỗi stack có Postgres riêng với 7 database/role: `littlecafe_<service>`.
- `database/init/00-init.sql` chỉ dùng cho local dev; staging/production dùng
  `deployments/scripts/init-db.sh` qua service `db-init` (idempotent).
- Migration Goose chạy qua service `migrate-<service>` với
  `command: ["-migrate-only"]` (không lặp `/app/app`).
- Migration phải chạy được trên database rỗng, idempotent và không đọc database
  của service khác.
- Không kết nối chéo môi trường (local → staging, staging → production).
- Thay đổi phá hủy schema phải có backup trước; ưu tiên forward-fix hơn rollback.

## 5. Provider bên thứ ba

| Provider | Local | Staging | Production |
| --- | --- | --- | --- |
| Google OAuth | trống | test client | client thật |
| payOS | trống | sandbox | thật |
| Brevo email | `EMAIL_ENABLED=false` | sandbox hoặc tắt | key thật |
| Cloudinary | trống | test account | thật |

- Không dùng credential production ở staging/local.
- Callback, return, cancel và webhook URL phải trỏ đúng môi trường:
  - Staging: `https://staging-api.littlecafe.kids:8443/...`
  - Production: `https://api.littlecafe.kids:8444/...`
- Staging không gửi email thật và không tạo giao dịch thật.

## 6. Kiểm tra theo môi trường

### Local

- Áp dụng gate trong `AGENTS.md` theo phạm vi thay đổi (bao gồm codegen/drift
  check nếu đổi contract hoặc Wire).
- Kiểm tra Compose bằng `docker compose ... config` khi sửa file Compose.
- Chạy full stack khi thay đổi ảnh hưởng runtime; kiểm tra health và CORS.

### Staging

- Push `develop` để CD build, migrate, rollout, smoke.
- Kiểm tra container healthy, health qua nginx, CORS với origin localhost,
  log không có lỗi cấu hình.
- Frontend test trên staging trước khi xem xét production.

### Production

- Merge `develop` → `main`, tag `v*`, theo dõi CD.
- Backup database trước thay đổi rủi ro.
- Smoke health, kiểm tra container, log và tài nguyên sau rollout.
- Rollback bằng digest/tag cũ; không `docker compose down -v` trên production.

### Definition of Done cho thay đổi chạm môi trường

- Config/env/Compose/CI/docs đã cập nhật đồng bộ.
- Không có secret trong repo, log hoặc diff.
- Migration chạy được trên DB rỗng.
- Test/gate tương ứng đã chạy.
- Staging đã verify nếu thay đổi ảnh hưởng hành vi runtime.
- Ghi chú rollback nếu thay đổi có rủi ro.

## 7. Luồng deploy và rollback

```text
feature branch -> develop -> CD staging -> verify/UAT
              -> main -> tag v* -> CD production -> smoke
```

- Chỉ tag trên `main`.
- Rollback: dispatch lại tag/commit cũ hoặc sửa digest trong `.env.deploy` /
  `.env.staging.deploy` rồi `up -d <service>`.
- Không chạy `down -v` trên production; chỉ dùng cho staging khi reset có chủ ý.
- Chi tiết: `docs/operations/deployment.md`, `deployment-staging.md`,
  `certificates.md` và `local-docs/deploy/`.

## 8. Pitfall đã gặp

- **Credential trong URL**: base64 có `/` hoặc `+` phải được URL-encode khi build
  DSN (AMQP đã dùng `url.UserPassword`); không nối chuỗi thủ công.
- **`-migrate-only`**: image có `ENTRYPOINT ["/app/app"]`, nên Compose `command`
  chỉ được là `["-migrate-only"]`; nếu lặp `/app/app`, Go flag parser bỏ qua flag
  và app khởi động full, cần RabbitMQ/Redis.
- **SQL migration**: không để literal `\n` trong file `.sql`; psql meta-command
  là `\connect` (một backslash), không phải `\\connect`.
- **Quyền secret**: container distroless nonroot không đọc được thư mục `700` và
  key `600` của deploy; chạy service bằng `APP_UID/APP_GID` khớp user deploy.
- **Smoke từ GitHub runner**: firewall VPS có thể chặn IP runner ở port
  8443/8444; smoke public nên chạy từ VPS hoặc chấp nhận kiểm tra thủ công.
- **Nginx static upstream**: nếu service upstream chết, nginx có thể không start;
  kiểm tra container trước khi recreate nginx.

## 9. Checklist trước khi merge

- [ ] Giá trị đặt đúng config hay env.
- [ ] Cập nhật 3 file `.env*.example`.
- [ ] Cập nhật Compose dev/staging/prod nếu cần.
- [ ] Cập nhật CI Compose validation nếu cần.
- [ ] Cập nhật docs operations/environments.
- [ ] Không có secret trong code, log, diff.
- [ ] Migration chạy được trên DB rỗng.
- [ ] Test/gate tương ứng đã chạy.
- [ ] Staging đã verify (nếu ảnh hưởng runtime).

## 10. Liên kết

- `docs/operations/deployment.md`
- `docs/operations/deployment-staging.md`
- `docs/operations/certificates.md`
- `.specify/memory/constitution.md`
- `local-docs/deploy/` (runbook vận hành, ngoài repo)
- `littlecafe-docs/deployment/environments.md` (chính sách mức dự án)
