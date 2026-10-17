# LittleCafe Backend

Nền tảng backend contract-first cho LittleCafe, module path `github.com/hycat-team/littlecafe-backend`.

## Workspace

- Bảy service độc lập: Identity, Order, Activity, Wallet, AI, Notification, Reporting.
- Content là module của Activity service.
- `shared` chứa framework dùng chung; `contracts` là nguồn chuẩn OpenAPI, AsyncAPI/JSON Schema và protobuf.
- Một PostgreSQL instance với bảy database và role tách biệt.

| Service | HTTP | gRPC |
|---|---:|---:|
| Identity | 8101 | 9101 |
| Order | 8102 | 9102 |
| Activity | 8103 | 9103 |
| Wallet | 8104 | 9104 |
| AI | 8105 | - |
| Notification | 8106 | - |
| Reporting | 8107 | - |

Nginx local dùng port 8080. Production dùng `https://api.littlecafe.kids/api/v1`; website chuẩn là `https://littlecafe.kids`.

## Khởi động

```powershell
Copy-Item .env.example .env
docker compose -f deployments/docker-compose.yml --profile infra up -d
make generate
make test
make build-all
```

Full stack yêu cầu certificate local tại `CERTS_DIR` nằm ngoài repository, sau đó chạy profile `full`. Không có API nghiệp vụ mẫu; bootstrap chỉ cung cấp health/readiness và nền hạ tầng.

Xem [bản đồ tài liệu](docs/README.md) và [quy ước contract](docs/api/contracts.md).

