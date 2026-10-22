# Triển khai

Production và staging chạy bằng Docker Compose trên một VPS. CI kiểm tra source,
contract, dependency boundary, image `linux/amd64` và Compose config; CI không
push image và không chạm môi trường chạy.

CD cho VPS:

- `.github/workflows/cd.yml` — production, chạy khi push tag `v*` hoặc dispatch.
- `.github/workflows/cd-staging.yml` — staging, chạy khi push `develop` hoặc
  dispatch.

Khi dispatch, input `services` nhận `all` hoặc danh sách phân tách bằng dấu phẩy,
ví dụ `identity,order`. Mặc định nên deploy `all` vì các service có quan hệ gRPC
và Nginx phụ thuộc toàn bộ stack. Deploy lẻ chỉ dùng khi thay đổi service độc lập
đã được xác nhận.

## 1. Topology trên VPS dùng chung

VPS có thể đang chạy một stack khác đã chiếm 80/443. littlecafe dùng port riêng
và tách hoàn toàn volume/network/secret:

| Stack | Thư mục | HTTP | HTTPS |
|---|---|---:|---:|
| Staging | `/opt/littlecafe-staging` | 8081 | 8443 |
| Production | `/opt/littlecafe` | 8082 | 8444 |

Không publish port Postgres/Redis/RabbitMQ. TLS do littlecafe tự quản trong
`TLS_DIR` (mặc định `/opt/littlecafe-tls`), nginx mount read-only; cert cấp/gia
hạn độc lập bằng lego qua DNS-01 của Vercel. Chi tiết:
[`certificates.md`](certificates.md).

## 2. Luồng CD

```text
tag/manual dispatch (prod) hoặc push develop (staging)
  -> plan + validate service list
  -> build/push image matrix lên GHCR
  -> resolve digest immutable
  -> SSH upload compose/nginx/scripts files
  -> db-init (idempotent) provision role/database/extension
  -> update .env.deploy / .env.staging.deploy
  -> compose pull -> migration -> up từng service
  -> recreate nginx
  -> smoke test container health + public health
```

## 3. GitHub Environment và secrets

Tạo hai environment `production` và `staging`, bật Required reviewers nếu muốn
có approval gate.

Required repository/environment secrets:

| Secret | Vai trò |
|---|---|
| `DEPLOY_HOST` | Hostname/IP VPS |
| `DEPLOY_USER` | SSH user |
| `DEPLOY_PORT` | SSH port |
| `DEPLOY_SSH_KEY` | Private key deploy |
| `DEPLOY_KNOWN_HOSTS` | Output đã kiểm tra từ `ssh-keyscan` |
| `GHCR_USERNAME` | Tài khoản đọc GHCR trên VPS |
| `GHCR_READ_TOKEN` | PAT có quyền `read:packages` |

Repository variables tùy chọn:

| Variable | Vai trò |
|---|---|
| `PRODUCTION_HEALTH_URL` | URL smoke production (mặc định `https://api.littlecafe.kids:8444/api/v1/health`) |
| `STAGING_HEALTH_URL` | URL smoke staging (mặc định `https://staging-api.littlecafe.kids:8443/api/v1/health`) |

File `/opt/littlecafe/.env.production` và `/opt/littlecafe-staging/.env.staging`
được tạo ngoài Git, chứa database credentials, provider keys, JWT paths và
đường dẫn secret mount. Compose nạp file này qua `env_file`; biến bắt buộc khai
báo `${VAR:?...}` nên thiếu giá trị sẽ fail ngay khi parse compose. Không đưa
các giá trị này vào GitHub workflow.

Workflow tự quản lý `.env.deploy` (prod) và `.env.staging.deploy` (staging).
File chỉ chứa image reference đã pin digest, ví dụ:

```text
IDENTITY_IMAGE=ghcr.io/hycat-team/littlecafe-backend-identity@sha256:<digest>
```

Không dùng `docker compose down -v` trên production; thao tác đó xóa volumes
database. Backup phải bao phủ cả bảy database trước thay đổi hạ tầng lớn.

## 4. Provision database

Bước đầu tiên của mỗi lần deploy chạy service `db-init` (profile `init`) với
`deployments/scripts/init-db.sh`. Script idempotent: tạo role
`littlecafe_<service>`, database `littlecafe_<service>`, extension `pgcrypto`
cho cả bảy database và `vector` cho `littlecafe_ai`. Password lấy từ
`<SERVICE>_DB_PASSWORD` trong env file tương ứng.

Không tạo bảng ở bước này. Sau `db-init`, CD chạy job `migrate-<service>`
(`--profile migrate`) trước rollout service; migration Goose của service mới
tạo schema.

Nếu cần provision thủ công khi `db-init` lỗi:

```bash
cd /opt/littlecafe
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml up -d postgres
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml --profile init run --rm db-init
```

## 5. Certificate

Nginx mount `TLS_DIR` (mặc định `/opt/littlecafe-tls`) read-only tại
`/etc/nginx/tls`; cert phải tồn tại trước khi nginx khởi động. Cấp cert lần đầu
bằng `deployments/scripts/issue-certs.sh` (lego, DNS-01 Vercel); gia hạn tự động
bằng `deployments/scripts/renew-certs.sh` qua cron. Không phụ thuộc stack khác.
Chi tiết: [`certificates.md`](certificates.md).

## 6. Production

### Deploy

Push tag `v*` hoặc dispatch `cd.yml`. Workflow build image tag release, pin
digest vào `.env.deploy`, chạy `db-init`, migration rồi rollout từng service.

### Kiểm tra

```bash
curl -fsS https://api.littlecafe.kids:8444/api/v1/health
```

Kiểm tra container, log, tài nguyên và CORS theo `FRONTEND_ORIGINS`.

### Rollback

Dispatch lại workflow với commit/tag chứa image mong muốn, hoặc sửa image digest
tương ứng trong `.env.deploy` rồi chạy:

```bash
cd /opt/littlecafe
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml up -d identity
```

Sau rollback:

```bash
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml ps
curl -fsS https://api.littlecafe.kids:8444/api/v1/health
```

## 7. Staging

Staging phục vụ frontend test API mà không cần chạy backend cùng hạ tầng ở máy
local. Stack chạy độc lập trên VPS, tách khỏi production và stack khác đang chiếm
80/443.

- Nhánh trigger: `develop` (push) hoặc dispatch.
- Thư mục VPS: `/opt/littlecafe-staging`.
- URL: `https://staging-api.littlecafe.kids:8443/api/v1`.
- CORS đọc `FRONTEND_ORIGINS` (CSV); mặc định staging cho
  `http://localhost:3000` và `http://127.0.0.1:3000`.

### Chuẩn bị một lần trên VPS

1. Tạo thư mục secret ngoài thư mục deploy (để `rotate-mtls-certs.sh` chấp nhận):

   ```bash
   sudo mkdir -p /opt/littlecafe-secrets/{certs,jwt,acme} /opt/littlecafe-staging-secrets/{certs,jwt} /opt/littlecafe-tls
   sudo chown -R deploy:deploy /opt/littlecafe-secrets /opt/littlecafe-staging-secrets /opt/littlecafe-tls
   ```

2. Tạo `/opt/littlecafe-staging/.env.staging` từ `.env.staging.example`, thay
   toàn bộ mật khẩu và provider key bằng giá trị thật; `chmod 600`.

3. Sinh mTLS cert riêng cho staging (chạy từ `/opt/littlecafe`):

   ```bash
   cd /opt/littlecafe
   sh deployments/scripts/rotate-mtls-certs.sh /opt/littlecafe-staging-secrets/certs
   ```

4. Sinh cặp JWT ES256 riêng cho staging trong
   `/opt/littlecafe-staging-secrets/jwt`:

   ```bash
   cd /opt/littlecafe-staging-secrets/jwt
   openssl ecparam -name prime256v1 -genkey -noout -out jwt-private.pem
   openssl ec -in jwt-private.pem -pubout -out jwt-public.pem
   ```

5. Thêm DNS A record trên Vercel: `staging-api` → IP VPS.

6. Tạo Vercel API token (quyền quản lý DNS `littlecafe.kids`) và lưu vào
   `/opt/littlecafe-secrets/vercel.token` (`chmod 600`). Cấp cert lần đầu bằng
   lego qua DNS-01:

   ```bash
   printf '%s' '<vercel-token>' > /opt/littlecafe-secrets/vercel.token
   printf '%s' '<email>' > /opt/littlecafe-secrets/acme/account.email
   chmod 600 /opt/littlecafe-secrets/vercel.token /opt/littlecafe-secrets/acme/account.email

   cd /opt/littlecafe
   sh deployments/scripts/issue-certs.sh issue
   ```

   Kiểm tra `/opt/littlecafe-tls/staging-api.littlecafe.kids/` có
   `fullchain.pem` và `privkey.pem`.

7. Mở firewall port `8443` (và `8081`, `8444`, `8082` nếu cần) ở firewall của
   nhà cung cấp VPS. Docker bypass ufw và `INPUT` thường đang `ACCEPT`, nên ưu
   tiên mở ở panel provider.

8. Cài cron gia hạn cert:

   ```cron
   0 3 * * * sh /opt/littlecafe/deployments/scripts/renew-certs.sh >> /home/deploy/littlecafe-renew.log 2>&1
   ```

### Deploy

Push lên `develop` để chạy `cd-staging.yml`, hoặc dispatch thủ công. Workflow
build image tag `staging-<sha>`, pin digest vào `.env.staging.deploy`, chạy
`db-init`, migration rồi rollout từng service.

### Kiểm tra

```bash
curl -fsS https://staging-api.littlecafe.kids:8443/api/v1/health

curl -i -X OPTIONS https://staging-api.littlecafe.kids:8443/api/v1/health \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET"
```

Phải thấy `Access-Control-Allow-Origin: http://localhost:3000`. Kiểm tra tài
nguyên để tránh ảnh hưởng stack khác:

```bash
docker stats --no-stream
free -h
df -h
```

### Rollback

Sửa digest trong `/opt/littlecafe-staging/.env.staging.deploy` về image cũ rồi:

```bash
cd /opt/littlecafe-staging
docker compose --env-file .env.staging --env-file .env.staging.deploy \
  -f deployments/docker-compose.staging.yml up -d identity
```

## 8. Runbook ngoài repo

Bộ runbook đầy đủ nằm ngoài repo tại `local-docs/deploy/` (variables,
main-process, operations, troubleshooting). Branch/release flow:
[`../development/git-workflow.md`](../development/git-workflow.md).
