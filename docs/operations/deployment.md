# Triển khai

Production chạy trên một VPS bằng Docker Compose. CI kiểm tra source, contract,
dependency boundary, image `linux/amd64` và Compose config; CI không push
image và không chạm production.

CD duy nhất cho VPS là `.github/workflows/cd.yml`. Workflow chạy khi push tag
`v*` hoặc khi được dispatch thủ công. Khi dispatch, input `services` nhận `all`
hoặc danh sách phân tách bằng dấu phẩy, ví dụ `identity,order`.

## Luồng CD

```text
tag/manual dispatch
  -> plan + validate service list
  -> build/push image matrix lên GHCR
  -> resolve digest immutable
  -> SSH upload compose/nginx files
  -> update /opt/littlecafe/.env.deploy
  -> compose pull -> migration -> up từng service
  -> recreate nginx
  -> smoke test public health
```

Mặc định nên deploy `all`, vì các service có quan hệ gRPC và Nginx phụ thuộc
toàn bộ stack. Deploy lẻ chỉ dùng khi thay đổi service độc lập đã được xác nhận.

## GitHub Environment và secrets

Tạo environment `production`, bật Required reviewers nếu muốn có approval gate.

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

File `/opt/littlecafe/.env.production` được tạo và quản lý ngoài Git, chứa
database credentials, provider keys, JWT keys và đường dẫn secret mount. Compose
nạp file này vào mọi service qua `env_file`; biến bắt buộc khai báo `${VAR:?...}`
nên thiếu giá trị sẽ fail ngay khi parse compose. Không đưa các giá trị này vào
GitHub workflow.

Workflow tự quản lý `/opt/littlecafe/.env.deploy`. File này chỉ chứa các image
reference đã pin digest, ví dụ:

```text
IDENTITY_IMAGE=ghcr.io/hycat-team/littlecafe-backend-identity@sha256:<digest>
```

Không dùng `docker compose down -v` trên production; thao tác đó xóa volumes
database. Backup phải bao phủ cả bảy database trước các thay đổi hạ tầng lớn.

## Provision database lần đầu

`database/init/00-init.sql` chỉ là bootstrap development (compose dev mount vào
`/docker-entrypoint-initdb.d`). Compose production không mount file này, nên lần
đầu dựng VPS phải tạo role, database và extension **một lần** bằng tài khoản admin
trước khi deploy service.

Điều kiện: `/opt/littlecafe/.env.production` đã có `POSTGRES_ADMIN_USER`,
`POSTGRES_ADMIN_PASSWORD` và password mạnh cho từng service
(`<SERVICE>_DB_PASSWORD`). `DB_USER`/`DB_NAME` cố định trong compose
(`littlecafe_<service>`). Password sinh ngoài Git, không commit.

```bash
cd /opt/littlecafe
docker compose --env-file .env.production -f deployments/docker-compose.prod.yml up -d postgres
docker compose --env-file .env.production -f deployments/docker-compose.prod.yml \
  exec postgres sh -c 'psql -U "$POSTGRES_USER" -d postgres'
```

Trong phiên `psql`, chạy SQL tương đương `database/init/00-init.sql` nhưng với
password thật lấy từ `.env.production` (dán trực tiếp vào terminal, không lưu vào
file trong repo). Với mỗi service trong bảy service: `CREATE ROLE
littlecafe_<service> LOGIN PASSWORD '...'` rồi `CREATE DATABASE
littlecafe_<service> OWNER littlecafe_<service>`. Sau đó `\connect` vào từng
database để `CREATE EXTENSION IF NOT EXISTS pgcrypto`; riêng `littlecafe_ai` thêm
`CREATE EXTENSION IF NOT EXISTS vector`.

- Chỉ chạy trên volume mới: `CREATE ROLE`/`CREATE DATABASE` lỗi nếu đã tồn tại.
- Không tạo bảng ở bước này. CD chạy job `migrate-<service>` (`--profile migrate`)
  trước rollout service và migration Goose của service mới tạo schema.
- Đổi password về sau: `ALTER ROLE ... PASSWORD`, cập nhật `.env.production`,
  rồi recreate service tương ứng.

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
curl -fsS https://api.littlecafe.kids/api/v1/health
```
