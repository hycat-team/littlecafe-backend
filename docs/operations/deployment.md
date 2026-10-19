# Triển khai

Production chạy bằng Docker Compose trên một VPS. CI kiểm tra source, contract,
dependency boundary, image `linux/amd64` và Compose config; CI không push
image và không chạm production.

CD cho VPS là `.github/workflows/cd.yml` (production) và
`.github/workflows/cd-staging.yml` (staging). Production chạy khi push tag
`v*` hoặc dispatch thủ công; staging chạy khi push nhánh `develop` hoặc
dispatch thủ công. Khi dispatch, input `services` nhận `all` hoặc danh sách
phân tách bằng dấu phẩy, ví dụ `identity,order`.

## Topology trên VPS dùng chung

VPS có thể đang chạy một stack khác đã chiếm 80/443.
littlecafe dùng port riêng và tách hoàn toàn volume/network/secret:

| Stack | Thư mục | HTTP | HTTPS |
|---|---|---:|---:|
| Staging | `/opt/littlecafe-staging` | 8081 | 8443 |
| Production | `/opt/littlecafe` | 8082 | 8444 |

Không publish port Postgres/Redis/RabbitMQ. TLS do littlecafe tự quản trong
`TLS_DIR` (mặc định `/opt/littlecafe-tls`), nginx mount read-only; cert cấp/gia
hạn độc lập bằng lego qua DNS-01 của Vercel (xem
[deployment-staging.md](deployment-staging.md)).

## Luồng CD

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

Mặc định nên deploy `all`, vì các service có quan hệ gRPC và Nginx phụ thuộc
toàn bộ stack. Deploy lẻ chỉ dùng khi thay đổi service độc lập đã được xác nhận.

## GitHub Environment và secrets

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
| `PRODUCTION_HEALTH_URL` | URL smoke test production (mặc định `https://api.littlecafe.kids:8444/api/v1/health`) |
| `STAGING_HEALTH_URL` | URL smoke test staging (mặc định `https://staging-api.littlecafe.kids:8443/api/v1/health`) |

File `/opt/littlecafe/.env.production` và `/opt/littlecafe-staging/.env.staging`
được tạo ngoài Git, chứa database credentials, provider keys, JWT paths và
đường dẫn secret mount. Compose nạp file này vào mọi service qua `env_file`;
biến bắt buộc khai báo `${VAR:?...}` nên thiếu giá trị sẽ fail ngay khi parse
compose. Không đưa các giá trị này vào GitHub workflow.

Workflow tự quản lý `.env.deploy` (prod) và `.env.staging.deploy` (staging).
File chỉ chứa image reference đã pin digest, ví dụ:

```text
IDENTITY_IMAGE=ghcr.io/hycat-team/littlecafe-backend-identity@sha256:<digest>
```

Không dùng `docker compose down -v` trên production; thao tác đó xóa volumes
database. Backup phải bao phủ cả bảy database trước các thay đổi hạ tầng lớn.

## Provision database

Bước đầu tiên của mỗi lần deploy chạy service `db-init` (profile `init`) với
`deployments/scripts/init-db.sh`. Script idempotent: tạo role
`littlecafe_<service>`, database `littlecafe_<service>`, extension `pgcrypto`
cho cả bảy database và `vector` cho `littlecafe_ai`. Password lấy từ
`<SERVICE>_DB_PASSWORD` trong env file tương ứng.

Không tạo bảng ở bước này. Sau `db-init`, CD chạy job `migrate-<service>`
(`--profile migrate`) trước rollout service và migration Goose của service mới
tạo schema.

Nếu cần provision thủ công khi `db-init` lỗi:

```bash
cd /opt/littlecafe
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml up -d postgres
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml --profile init run --rm db-init
```

## Certificate

Nginx mount `TLS_DIR` (mặc định `/opt/littlecafe-tls`) read-only tại
`/etc/nginx/tls`; cert phải tồn tại trước khi nginx khởi động. Cấp cert lần đầu
bằng `deployments/scripts/issue-certs.sh` (lego, DNS-01 Vercel); gia hạn tự động
bằng `deployments/scripts/renew-certs.sh` qua cron. Không phụ thuộc stack khác.
Chi tiết trong [deployment-staging.md](deployment-staging.md) và
[certificates.md](certificates.md).

## Rollback

Rollback service bằng cách dispatch lại workflow với commit/tag chứa image mong
muốn, hoặc sửa image digest tương ứng trong `.env.deploy` rồi chạy:

```bash
cd /opt/littlecafe
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml up -d identity
```

Sau rollback phải kiểm tra:

```bash
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml ps
curl -fsS https://api.littlecafe.kids:8444/api/v1/health
```
