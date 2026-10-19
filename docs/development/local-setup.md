# Phát triển local

Trước khi thêm hoặc sửa config, secret hay hành vi phụ thuộc môi trường, đọc
[environments.md](environments.md).

1. Copy `.env.example` thành `.env` và giữ file thật ngoài Git. Compose nạp `.env` vào mọi container qua `env_file`, nên biến provider (Google, payOS, Brevo, Cloudinary, JWT...) phải nằm trong file này, không chỉ để interpolate.
2. Tạo certificate local ở thư mục nằm ngoài repository bằng `deployments/scripts/rotate-mtls-certs.sh`; `CERTS_DIR` trong `.env` phải trỏ tới thư mục đó.
3. Chạy `docker compose -f deployments/docker-compose.yml --profile infra up -d` để bật PostgreSQL, Redis và RabbitMQ.
4. Chạy `make generate`, `make test`, `make build-all`.
5. Chạy full stack bằng profile `full` sau khi certificate đã sẵn sàng. Port host cố định: HTTP 8101-8107, gRPC 9101-9104, Nginx 8080.

Không thêm entity/endpoint mẫu. Mỗi thay đổi nghiệp vụ bắt đầu từ requirement, workflow, ADR và contract đã duyệt.

