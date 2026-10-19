# Triển khai staging

Staging phục vụ frontend test API mà không cần chạy backend cùng hạ tầng ở máy
local. Stack chạy độc lập trên VPS, tách khỏi production và khỏi stack khác
đang chiếm 80/443.

- Nhánh trigger: `develop` (push) hoặc dispatch thủ công.
- Thư mục trên VPS: `/opt/littlecafe-staging`.
- URL: `https://staging-api.littlecafe.kids:8443/api/v1`.
- CORS cho phép `http://localhost:*` và `http://127.0.0.1:*`.

## Chuẩn bị một lần trên VPS

1. Tạo thư mục secret **ngoài thư mục deploy** (để `rotate-mtls-certs.sh` chấp
   nhận):

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
   tiên mở ở panel provider; nếu provider không có firewall thì mặc định đã mở.

8. Cài cron gia hạn cert:

   ```cron
   0 3 * * * sh /opt/littlecafe/deployments/scripts/renew-certs.sh >> /home/deploy/littlecafe-renew.log 2>&1
   ```

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
