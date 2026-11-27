# Triển khai

Canonical cho: CD, deploy, rollback trên VPS. Env/key:
[`../development/sops-keys.md`](../development/sops-keys.md); giá trị config:
[`../development/environments.md`](../development/environments.md); certificate:
[`certificates.md`](certificates.md).

Production và staging chạy Docker Compose trên một VPS. CI kiểm tra source,
contract, boundary, image `linux/amd64` và Compose config; CI không push image và
không chạm môi trường chạy.

- `.github/workflows/cd.yml` — production, tag `v*` hoặc dispatch.
- `.github/workflows/cd-staging.yml` — staging, push `develop` hoặc dispatch.
- Dispatch nhận `services=all` hoặc CSV (`commerce,experience`). Mặc định `all`
  vì service có quan hệ gRPC và Nginx phụ thuộc toàn stack.

## 1. Topology trên VPS dùng chung

| Stack | Thư mục | HTTP | HTTPS |
|---|---|---:|---:|
| Staging | `/opt/littlecafe-staging` | 8081 | 8443 |
| Production | `/opt/littlecafe` | 8082 | 8444 |

- Không publish Postgres/Redis/RabbitMQ ra public; ngoại lệ có kiểm soát:
  Postgres bind loopback `127.0.0.1:5433` (prod), `127.0.0.1:5434` (staging).
- TLS do littlecafe tự quản trong `TLS_DIR` (`/opt/littlecafe-tls`), nginx mount
  read-only; cấp/gia hạn bằng lego DNS-01 Vercel. Chi tiết:
  [`certificates.md`](certificates.md).
- VPS có thể chạy stack khác; chỉ thao tác trong `/opt/littlecafe*` và container
  prefix `littlecafe-`.

## 2. Luồng CD

```text
tag/dispatch (prod) hoặc push develop (staging)
  -> plan + validate service list
  -> build/push image matrix lên GHCR
  -> SSH upload compose/nginx/scripts/env ciphertext
  -> resolve digest immutable cho service được chọn
  -> apply-env.sh decrypt SOPS (một lần mỗi stack)
  -> deploy-release.sh (một SSH, một writer .env.deploy):
       - backup .env.deploy -> .env.deploy.prev
       - ghi image digest atomic
       - compose pull service được chọn
       - up -d --wait postgres redis rabbitmq
       - db-init (idempotent)
       - migrate-<service> song song, fail-fast
       - up -d --wait --no-deps service
       - recreate nginx sau cùng
  -> smoke container health + public health
```

`depends_on` chỉ giữ Postgres cho service app vì Redis/RabbitMQ có startup retry;
Nginx recreate sau khi service healthy. Local compose giữ `depends_on` đầy đủ.

## 3. GitHub Environment, secret và env file

Tạo environment `production`/`staging`, bật Required reviewers nếu cần approval.

| Secret | Vai trò |
|---|---|
| `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_PORT` | Kết nối SSH |
| `DEPLOY_SSH_KEY`, `DEPLOY_KNOWN_HOSTS` | Private key + host key đã kiểm tra |
| `GHCR_USERNAME`, `GHCR_READ_TOKEN` | Đọc GHCR trên VPS |

Variable tùy chọn: `PRODUCTION_HEALTH_URL`, `STAGING_HEALTH_URL`.

API docs staging dùng basic auth; credential nằm trong SOPS staging env
(`DOCS_BASIC_AUTH_USER`, `DOCS_BASIC_AUTH_PASSWORD`), không phải GitHub secret.

Env staging/production là ciphertext SOPS trong Git; CD giải mã trên VPS thành
`.env.staging`/`.env.production` (mode 600). Compose nạp qua `env_file`; biến bắt
buộc dùng `${VAR:?...}` nên thiếu giá trị fail ngay khi parse. Không đưa secret
ứng dụng vào GitHub workflow.

- Age identity: `/opt/littlecafe-secrets/age/keys.txt` (prod),
  `/opt/littlecafe-staging-secrets/age/keys.txt` (staging), 600, không mount.
- Plaintext trên VPS là artifact dẫn xuất; sửa qua Git
  (`make env-edit`) rồi deploy. Quy trình/rotation: `sops-keys.md`.
- `.env.deploy` (prod) / `.env.staging.deploy` (staging) do CD quản lý, chỉ chứa
  image digest:
  `COMMERCE_IMAGE=ghcr.io/hycat-team/littlecafe-backend-commerce@sha256:<digest>`.
- Không `docker compose down -v` (xóa volume DB); backup cả 4 database trước thay
  đổi hạ tầng lớn.

## 4. Provision database

Sau khi `postgres redis rabbitmq` healthy, CD chạy `db-init` (profile `init`) với
`deployments/scripts/init-db.sh`: tạo role/database `littlecafe_<service>`,
extension `pgcrypto` (4 DB) và `vector` (AI). Password lấy từ
`<SERVICE>_DB_PASSWORD`.

Sau `db-init`, `migrate-<service>` (profile `migrate`) chạy song song trước
rollout; Goose tạo schema trong DB riêng. Nếu `db-init` lỗi, chạy thủ công:

```bash
cd /opt/littlecafe
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml up -d postgres
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml --profile init run --rm db-init
```

## 5. Certificate

Nginx mount `TLS_DIR` read-only tại `/etc/nginx/tls`; cert phải tồn tại trước khi
nginx start. Cấp lần đầu bằng `issue-certs.sh`, gia hạn bằng `renew-certs.sh` qua
cron. Chi tiết cert/JWT/mTLS: [`certificates.md`](certificates.md).

## 6. Production

### Deploy

Push tag `v*` hoặc dispatch `cd.yml`. Workflow build image release, pin digest vào
`.env.deploy`, rồi `deploy-release.sh` pull, `db-init`, migrate song song, rollout
4 service, recreate Nginx. App health fail sau rollout → script khôi phục
`.env.deploy.prev` và rollback image cũ.

### Kiểm tra

```bash
curl -fsS https://api.littlecafe.kids:8444/api/v1/health
```

Kiểm tra container, log, tài nguyên, CORS theo `FRONTEND_ORIGINS`.

### Rollback

Dispatch lại workflow với commit/tag cũ, hoặc:

```bash
cd /opt/littlecafe
cp -a .env.deploy.prev .env.deploy
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml up -d --wait --no-deps \
  commerce experience ai platform
docker compose --env-file .env.production --env-file .env.deploy \
  -f deployments/docker-compose.prod.yml up -d --force-recreate --no-deps nginx
```

Migration forward-only; rollback chỉ đổi image, không rollback schema. Sau đó
kiểm tra `docker compose ... ps` và health. Env rollback: xem
[`sops-keys.md`](../development/sops-keys.md) §5.6.

## 7. Staging

- Trigger: push `develop` hoặc dispatch. Thư mục `/opt/littlecafe-staging`.
- URL `https://staging-api.littlecafe.kids:8443/api/v1`; CORS mặc định
  `http://localhost:3000`, `http://127.0.0.1:3000`.

### Chuẩn bị một lần

1. Thư mục + age key: xem [`certificates.md`](certificates.md) §JWT/TLS và
   [`sops-keys.md`](../development/sops-keys.md) §6.4.
2. mTLS/JWT riêng cho staging: `certificates.md`.
3. DNS A record `staging-api` → IP VPS; mở firewall 8443 (và 8081/8444/8082 nếu
   cần) ở panel provider (Docker bypass ufw).
4. Cấp cert lần đầu + cron gia hạn: `certificates.md` và `issue-certs.sh`/
   `renew-certs.sh`.

### Deploy và kiểm tra

Push `develop` → `cd-staging.yml`; workflow build `staging-<sha>`, pin digest,
`db-init`, migrate, rollout. Kiểm tra:

```bash
curl -fsS https://staging-api.littlecafe.kids:8443/api/v1/health
curl -i -X OPTIONS https://staging-api.littlecafe.kids:8443/api/v1/health \
  -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: GET"
```

Phải thấy `Access-Control-Allow-Origin: http://localhost:3000`. Kiểm tra
`docker stats --no-stream`, `free -h`, `df -h` để tránh ảnh hưởng stack khác.

### API docs (staging)

- URL: `https://staging-api.littlecafe.kids:8443/docs/` (basic auth).
- Credential: `DOCS_BASIC_AUTH_USER` / `DOCS_BASIC_AUTH_PASSWORD` trong
  `deployments/env/staging.env.enc`; CD render `docs-auth/docs.htpasswd` trên VPS.
- Build: `cd-staging.yml` chạy `make docs-build` với server URL staging rồi mount
  `docs-site/` vào nginx. Docs chỉ có ở staging; production trả 404.
- Rotation: `make env-edit ENV=staging` → đổi password → re-run `CD Staging`.

### Rollback

Sửa digest trong `/opt/littlecafe-staging/.env.staging.deploy` về image cũ rồi
recreate service:

```bash
cd /opt/littlecafe-staging
docker compose --env-file .env.staging --env-file .env.staging.deploy \
  -f deployments/docker-compose.staging.yml up -d commerce
```

## 8. Runbook ngoài repo

Runbook đầy đủ: runbook nội bộ (ngoài repo, local-only; variables, bootstrap, operations,
troubleshooting). Branch/release: [`../development/git-workflow.md`](../development/git-workflow.md).
