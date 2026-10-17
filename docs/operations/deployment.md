# Triển khai

Production chạy trên một VPS bằng Docker Compose. CI kiểm tra source, contract,
dependency boundary, binary, image `linux/amd64` và Compose config; CI không push
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
database credentials, provider keys, JWT keys và đường dẫn secret mount. Không
đưa các giá trị này vào GitHub workflow.

Workflow tự quản lý `/opt/littlecafe/.env.deploy`. File này chỉ chứa các image
reference đã pin digest, ví dụ:

```text
IDENTITY_IMAGE=ghcr.io/hycat-team/littlecafe-backend-identity@sha256:<digest>
```

Không dùng `docker compose down -v` trên production; thao tác đó xóa volumes
database. Backup phải bao phủ cả bảy database trước các thay đổi hạ tầng lớn.

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
