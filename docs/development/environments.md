# Môi trường phát triển và triển khai

Canonical cho: phân loại config/env/secret, pitfall môi trường. Local setup:
[`bootstrap.md`](bootstrap.md). Deploy/rollback: [`../operations/deployment.md`](../operations/deployment.md);
cập nhật env và age key: [`sops-keys.md`](sops-keys.md); policy mức dự án:
`littlecafe-docs/deployment/environments.md`.

## 1. Bảng môi trường

| Môi trường | Mục đích | Trigger | URL | Thư mục VPS | Compose |
| --- | --- | --- | --- | --- | --- |
| Local | Phát triển, test nhanh | bất kỳ | `http://localhost:8080` | — | `deployments/docker-compose.yml` |
| Staging | Frontend test, UAT | push `develop` / dispatch | `https://staging-api.littlecafe.kids:8443` | `/opt/littlecafe-staging` | `deployments/docker-compose.staging.yml` |
| Production | Người dùng thật | tag `v*` / dispatch | `https://api.littlecafe.kids:8444` | `/opt/littlecafe` | `deployments/docker-compose.prod.yml` |

- Cùng VPS, tách compose project, volume, network, secret, JWT và CA.
- 80/443 thuộc stack khác; littlecafe dùng 8081/8443 (staging), 8082/8444 (prod).
- Local không cần mTLS; chỉ cần JWT nếu test đăng nhập.

## 2. Config và biến môi trường

### 2.1. Đặt giá trị ở đâu

| Loại giá trị | Nơi đặt | Ví dụ |
| --- | --- | --- |
| Hằng số ứng dụng | `configs/base.yaml` hoặc `configs/<service>.yaml` | port, timeout, rate limit, batch size |
| Host/endpoint khác theo môi trường | env hoặc Compose `environment` | `DB_HOST`, `REDIS_HOST`, `RABBITMQ_HOST` |
| Secret | env ciphertext trong Git (staging/prod), `.env` local | DB password, provider key, key path |
| Flag vận hành | env | `APP_MODE`, `SERVER_ENV`, `LOG_LEVEL`, `EMAIL_ENABLED` |
| CORS frontend | `FRONTEND_ORIGINS` (CSV) | `https://a.example.com,https://b.example.com` |
| Chỉ Compose dùng | Compose hoặc env | `TZ`, mount path, `APP_UID`/`APP_GID` |

- Giá trị giống nhau ở mọi môi trường là literal trong `configs/`.
- Chỉ secret và giá trị thực sự khác môi trường mới vào env.
- Không hardcode host, port, URL, credential trong code.
- Không rẽ nhánh business logic theo môi trường; dùng config hoặc feature flag.
- Không `os.Getenv` trong `domain`/`application`/`presentation`; đi qua
  `shared/config`, validation fail fast khi khởi động.
- CORS xử lý ở Go middleware; `*` chỉ hợp lệ ngoài production.

### 2.2. Thêm hoặc đổi một setting

1. Phân loại theo bảng 2.1.
2. Hằng số → sửa `configs/`. Env/secret → cập nhật `.env.<env>.example` và
   ciphertext theo [`sops-keys.md`](sops-keys.md) §5.
3. Cập nhật Compose/CI nếu setting ảnh hưởng `environment`, `ports`, `volumes`,
   `command`, `user` hoặc validation Compose.
4. Chạy `docker compose config` cho cả 3 môi trường, `make env-check`; thêm test
   config loader nếu có ràng buộc.

## 3. Secret và file

- Env staging/production trong Git là ciphertext SOPS:
  `deployments/env/{staging,production}.env.enc`; không commit plaintext, private
  key age không nằm trong repo.
- Local: `.env` riêng từng dev (gitignored) sinh từ `make env-local`; không chứa
  secret thật nên không mã hóa.
- Plaintext runtime chỉ trên VPS (`/opt/littlecafe/.env.production`,
  `/opt/littlecafe-staging/.env.staging`, mode 600), do CD giải mã bằng
  `deployments/scripts/apply-env.sh`; không sửa tay.
- `.env*.example` chỉ chứa placeholder; comment trong `.env*` viết tiếng Anh.
- `.env.<env>.example` là template cho thứ tự key, section comment và blank line;
  `apply-env.sh` render `.env.<env>` theo template này. Comment/thứ tự trong
  `.enc` không điều khiển runtime.
- Vị trí secret trên VPS:
  - Production: `/opt/littlecafe-secrets/{certs,jwt,acme,age}` + `vercel.token`.
  - Staging: `/opt/littlecafe-staging-secrets/{certs,jwt,age}`.
  - TLS nginx: `/opt/littlecafe-tls/<domain>/{fullchain.pem,privkey.pem}`.
  - `age/` (700, key 600) chứa identity giải mã env, không mount vào container.
- `APP_UID`/`APP_GID` khớp user deploy (mặc định 1000) để container đọc secret
  700/600 mà không mở quyền toàn hệ thống.
- JWT ES256: mọi service mount public key; chỉ Commerce mount private key. mTLS:
  4 service gRPC cần server cert; `service-client` dùng khi gọi service khác.
- Cấp quyền/rotate/offboarding/recovery: [`sops-keys.md`](sops-keys.md).
- Token Vercel và GHCR read token chỉ nằm trên VPS/CI secret; không log secret.

## 4. Local setup

Checklist đầy đủ: [`bootstrap.md`](bootstrap.md) — canonical cho clone → env →
JWT → infra → full stack → mTLS/SOPS. Local không cần mTLS; JWT key nằm trong
`CERTS_DIR` ngoài repo.

Không thêm entity/endpoint mẫu; thay đổi nghiệp vụ bắt đầu từ requirement,
workflow, ADR và contract đã duyệt.

## 5. Database và migration

- Mỗi stack có Postgres riêng với 4 database/role: `littlecafe_<service>`.
- Local dùng `database/init/00-init.sql`; staging/production dùng
  `deployments/scripts/init-db.sh` qua service `db-init` (idempotent).
- Goose chạy qua `migrate-<service>` với `command: ["-migrate-only"]`.
- Migration phải chạy trên DB rỗng, idempotent, không đọc DB service khác; không
  kết nối chéo môi trường.
- Thay đổi phá hủy schema phải backup trước; ưu tiên forward-fix.
- Chi tiết schema/Goose: [`../system-design/data-and-migrations.md`](../system-design/data-and-migrations.md).

## 6. Provider bên thứ ba

| Provider | Local | Staging | Production |
| --- | --- | --- | --- |
| Google OAuth | trống | test client | client thật |
| payOS | trống | sandbox | thật |
| Brevo email | `EMAIL_ENABLED=false` | sandbox hoặc tắt | key thật |
| Cloudinary | trống | test account | thật |

- Không dùng credential production ở staging/local.
- Callback/return/cancel/webhook URL trỏ đúng môi trường.
- Staging không gửi email thật và không tạo giao dịch thật.

## 7. Kiểm tra theo môi trường

- Gate kỹ thuật: [`gates.md`](gates.md) — nguồn chuẩn duy nhất.
- Local: `docker compose ... config` khi sửa Compose; chạy full stack khi thay
  đổi runtime, kiểm tra health/CORS.
- Staging: push `develop`, kiểm tra container/health/CORS/log trước khi xét
  production. Production: merge `main` + tag/`dispatch`, backup DB nếu rủi ro,
  smoke health/container/log/tài nguyên.
- Không `docker compose down -v`; rollback theo
  [`../operations/deployment.md`](../operations/deployment.md) §6–§7.

## 8. Pitfall đã gặp

- **Credential trong URL**: base64 có `/` hoặc `+` phải URL-encode khi build DSN
  (AMQP dùng `url.UserPassword`).
- **`-migrate-only`**: image có `ENTRYPOINT ["/app/app"]`, Compose `command` chỉ
  được là `["-migrate-only"]`.
- **SQL migration**: không để literal `\n`; psql meta-command là `\connect`.
- **Quyền secret**: container distroless nonroot không đọc được thư mục 700 và key
  600; dùng `APP_UID/APP_GID` khớp user deploy.
- **Smoke từ GitHub runner**: firewall VPS có thể chặn IP runner ở 8443/8444.
- **Nginx static upstream**: upstream chết có thể làm nginx không start; kiểm tra
  container trước khi recreate nginx.

## 9. Liên kết

- Deploy/rollback: [`../operations/deployment.md`](../operations/deployment.md).
- Certificate: [`../operations/certificates.md`](../operations/certificates.md).
- Env/key: [`sops-keys.md`](sops-keys.md); branch/release:
  [`git-workflow.md`](git-workflow.md).
- Policy: `.specify/memory/constitution.md` và
  `littlecafe-docs/deployment/environments.md`; runbook: runbook nội bộ (ngoài repo, local-only).
