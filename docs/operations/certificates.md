# Certificate và private CA

Script `deployments/scripts/rotate-mtls-certs.sh` tạo CA riêng, certificate server cho Identity, Order, Activity, Wallet và certificate client. Bắt buộc truyền một đường dẫn tuyệt đối nằm ngoài repository; khi xoay vòng, bộ cũ được đổi tên thành bản backup có timestamp để có thể rollback:

```sh
sh deployments/scripts/rotate-mtls-certs.sh /opt/littlecafe-secrets/certs
sh deployments/scripts/rotate-mtls-certs.sh /opt/littlecafe-staging-secrets/certs
```

Thời hạn mặc định là 90 ngày cho leaf certificate và 10 năm cho CA; có thể chỉnh bằng `LITTLECAFE_CERT_DAYS` và `LITTLECAFE_CA_DAYS`. Sau khi rotate, restart tuần tự client rồi server và chỉ xóa backup sau khi xác nhận toàn bộ kết nối.

Repository chỉ định contract biến môi trường: `GRPC_TLS_CA_FILE`, `GRPC_TLS_CERT_FILE`, `GRPC_TLS_KEY_FILE`, `GRPC_MTLS_REQUIRED=true`.

Không commit private key. Production phải fail closed khi mTLS được yêu cầu mà thiếu một trong ba đường dẫn certificate.

## Service nào cần certificate nào

mTLS chỉ áp dụng cho giao tiếp **gRPC**. Hiện chỉ 4 service chạy gRPC server
(`internal/bootstrap/grpc.go` + `NewGRPCServer` + `grpcx.ServerOptionsFromEnv`):

| Service | gRPC server | Cert cần | Ghi chú |
| --- | --- | --- | --- |
| Identity | có | `identity.crt/.key` + `ca.crt` | thêm JWT private key để ký |
| Order | có | `order.crt/.key` + `ca.crt` | |
| Activity | có | `activity.crt/.key` + `ca.crt` | gồm cả Content |
| Wallet | có | `wallet.crt/.key` + `ca.crt` | |
| AI | không | `ca.crt` (chỉ khi gọi gRPC) | hiện chỉ HTTP + event |
| Notification | không | `ca.crt` (chỉ khi gọi gRPC) | hiện chỉ HTTP + event |
| Reporting | không | `ca.crt` (chỉ khi gọi gRPC) | hiện chỉ HTTP + event |

- `service-client.crt/.key`: **client cert** cho service đi gọi service khác qua
  gRPC (`grpcx.DialService`). Hiện chưa service nào gọi nên chưa dùng, nhưng đã
  cấp sẵn cho các gRPC read sắp tới.
- `ca.crt`: CA tin cậy, cần cho cả phía server lẫn client.
- JWT: mọi service cần `jwt-public.pem` để verify; chỉ Identity cần
  `jwt-private.pem` để ký (xem mục dưới).

### Khi thêm một service dùng gRPC

1. `deployments/scripts/rotate-mtls-certs.sh`: thêm tên service vào vòng lặp
   `for service in identity order activity wallet` để cấp server cert
   `<service>.crt/.key`. Sau khi sửa, copy script lên VPS (re-scp `deployments/`)
   rồi chạy lại cho từng môi trường (bộ cũ tự backup):

   ```bash
   cd /opt/littlecafe
   sh deployments/scripts/rotate-mtls-certs.sh /opt/littlecafe-secrets/certs
   sh deployments/scripts/rotate-mtls-certs.sh /opt/littlecafe-staging-secrets/certs
   ```

2. `deployments/docker-compose.prod.yml` và `deployments/docker-compose.staging.yml`:
   thêm `GRPC_TLS_CA_FILE`, `GRPC_TLS_CERT_FILE`, `GRPC_TLS_KEY_FILE` và mount
   `${CERTS_DIR}:/run/littlecafe/certs:ro` cho service đó.
3. Service cần `internal/bootstrap/grpc.go` dùng `grpcx.ServerOptionsFromEnv()` và
   được wire trong `internal/di`.
4. Nếu service chỉ **gọi** service khác (gRPC client), nó cần
   `service-client.crt/.key` + `ca.crt` và dùng `grpcx.DialService`; không cần
   server cert.
5. Xoay vòng định kỳ/đổi cert: chạy lại `rotate-mtls-certs.sh` như bước 1, rồi
   restart client trước, server sau; kiểm tra kết nối rồi mới xóa backup.

## JWT key và TLS

- JWT ES256 nằm trong `JWT_DIR` (mặc định `/opt/littlecafe-secrets/jwt` và
  `/opt/littlecafe-staging-secrets/jwt`). Chỉ Identity mount private key; các
  service khác chỉ mount `jwt-public.pem`.
- TLS của nginx đọc từ `TLS_DIR/<API_DOMAIN>/` (mặc định `/opt/littlecafe-tls`),
  mount read-only tại `/etc/nginx/tls`. Cert cấp/gia hạn độc lập bằng
  `deployments/scripts/issue-certs.sh` và `deployments/scripts/renew-certs.sh`
  (lego, DNS-01 Vercel), không phụ thuộc stack khác; xem
  [deployment-staging.md](deployment-staging.md).

