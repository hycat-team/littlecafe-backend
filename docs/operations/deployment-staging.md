# Triển khai staging

Staging phục vụ frontend test API mà không cần chạy backend cùng hạ tầng ở máy
local. Stack chạy độc lập trên VPS, tách khỏi production và khỏi stack khác
đang chiếm 80/443.

- Nhánh trigger: `develop` (push) hoặc dispatch thủ công.
- Thư mục trên VPS: `/opt/littlecafe-staging`.
- URL: `https://staging-api.littlecafe.kids:8443/api/v1`.
- CORS cho phép `http://localhost:*` và `http://127.0.0.1:*`.

## Chuẩn bị một lần trên VPS

1. Tạo thư mục và quyền:

   ```bash
   sudo mkdir -p /opt/littlecafe-staging/secrets/{certs,jwt}
   sudo chown -R deploy:deploy /opt/littlecafe-staging
   ```

2. Tạo `/opt/littlecafe-staging/.env.staging` từ `.env.staging.example`, thay
   toàn bộ mật khẩu và provider key bằng giá trị thật; `chmod 600`.

3. Sinh mTLS cert riêng cho staging:

   ```bash
   deployments/scripts/rotate-mtls-certs.sh /opt/littlecafe-staging/secrets/certs
   ```

4. Sinh cặp JWT ES256 riêng cho staging trong
   `/opt/littlecafe-staging/secrets/jwt`:

   ```bash
   cd /opt/littlecafe-staging/secrets/jwt
   openssl ecparam -name prime256v1 -genkey -noout -out jwt-private.pem
   openssl ec -in jwt-private.pem -pubout -out jwt-public.pem
   ```

5. Thêm DNS A record trên Vercel: `staging-api` → IP VPS.

6. Tạo Vercel API token (quyền quản lý DNS `littlecafe.kids`) và lưu vào
   `/opt/littlecafe/secrets/vercel.token` (`chmod 600`). Cấp cert lần đầu bằng
   lego qua DNS-01:

   ```bash
   sudo mkdir -p /etc/letsencrypt/live/staging-api.littlecafe.kids
   sudo chown -R deploy:deploy /etc/letsencrypt/live/staging-api.littlecafe.kids
   ACME_EMAIL=<email> deployments/scripts/issue-certs.sh issue
   ```

   Kiểm tra `/etc/letsencrypt/live/staging-api.littlecafe.kids/` có
   `fullchain.pem` và `privkey.pem`. Có thể chạy `ACME_EMAIL=<email>` trong
   cron bằng cách đặt biến trong script hoặc wrapper.

7. Mở firewall port `8443` (và `8081` nếu cần):
   `sudo ufw allow 8443/tcp`.

8. Cài cron gia hạn cert gọi `deployments/scripts/renew-certs.sh`.

## Deploy

Push lên `develop` để chạy `.github/workflows/cd-staging.yml`, hoặc dispatch thủ
công. Workflow build image tag `staging-<sha>`, pin digest vào
`.env.staging.deploy`, chạy `db-init`, migration rồi rollout từng service.

## Kiểm tra

```bash
curl -fsS https://staging-api.littlecafe.kids:8443/api/v1/health

curl -i -X OPTIONS https://staging-api.littlecafe.kids:8443/api/v1/health \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET"
```

Phải thấy `Access-Control-Allow-Origin: http://localhost:3000`.

Kiểm tra tài nguyên để tránh ảnh hưởng stack khác:

```bash
docker stats --no-stream
free -h
df -h
```

## Rollback

Sửa digest trong `/opt/littlecafe-staging/.env.staging.deploy` về image cũ rồi:

```bash
cd /opt/littlecafe-staging
docker compose --env-file .env.staging --env-file .env.staging.deploy \
  -f deployments/docker-compose.staging.yml up -d identity
```

## Runbook chi tiết

Bộ runbook đầy đủ nằm ngoài repo tại `local-docs/deploy/` (variables,
main-process, operations, troubleshooting).
