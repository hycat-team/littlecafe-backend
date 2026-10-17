# Phát triển local

1. Copy `.env.example` thành `.env` và giữ file thật ngoài Git.
2. Tạo certificate local ở thư mục nằm ngoài repository bằng `deployments/scripts/rotate-mtls-certs.sh`; `CERTS_DIR` trong `.env` phải trỏ tới thư mục đó.
3. Chạy `docker compose -f deployments/docker-compose.yml --profile infra up -d` để bật PostgreSQL, Redis và RabbitMQ.
4. Chạy `make generate`, `make test`, `make build-all`.
5. Chạy full stack bằng profile `full` sau khi certificate đã sẵn sàng.

Không thêm entity/endpoint mẫu. Mỗi thay đổi nghiệp vụ bắt đầu từ requirement, workflow, ADR và contract đã duyệt.

