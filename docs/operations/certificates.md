# Certificate và private CA

Script `deployments/scripts/rotate-mtls-certs.sh` tạo CA riêng, certificate server cho Identity, Order, Activity, Wallet và certificate client. Bắt buộc truyền một đường dẫn tuyệt đối nằm ngoài repository; khi xoay vòng, bộ cũ được đổi tên thành bản backup có timestamp để có thể rollback:

```sh
deployments/scripts/rotate-mtls-certs.sh /opt/littlecafe/secrets/certs
```

Thời hạn mặc định là 90 ngày cho leaf certificate và 10 năm cho CA; có thể chỉnh bằng `LITTLECAFE_CERT_DAYS` và `LITTLECAFE_CA_DAYS`. Sau khi rotate, restart tuần tự client rồi server và chỉ xóa backup sau khi xác nhận toàn bộ kết nối.

Repository chỉ định contract biến môi trường: `GRPC_TLS_CA_FILE`, `GRPC_TLS_CERT_FILE`, `GRPC_TLS_KEY_FILE`, `GRPC_MTLS_REQUIRED=true`.

Không commit private key. Production phải fail closed khi mTLS được yêu cầu mà thiếu một trong ba đường dẫn certificate.

