# Bootstrap dự án

Entry point khi clone repository. Làm tuần tự các bước dưới; chi tiết sâu nằm ở
các file canonical được link ở từng bước.

## 1. Prerequisites

- Go `>= 1.26`
- Docker + Docker Compose
- GNU Make; Windows dùng Git Bash/Chocolatey
- `openssl` (Git Bash có sẵn) để tạo JWT key local
- `make install-tools` khi cần đổi protobuf (Buf, Wire)

## 2. Env local

```powershell
make env-local          # copy .env.example -> .env, không ghi đè
make env-local-check    # so key với .env.example
```

`.env` là per-dev, không commit. Giá trị mặc định dùng DB/RabbitMQ password
`12345678`, Redis không password, provider tắt.

## 3. JWT keys

JWT ES256 cần cho login local; key nằm trong `CERTS_DIR` (ngoài repo):

```bash
mkdir -p ../littlecafe-secrets/certs
openssl ecparam -name prime256v1 -genkey -noout \
  -out ../littlecafe-secrets/certs/jwt-private.pem
openssl ec -in ../littlecafe-secrets/certs/jwt-private.pem -pubout \
  -out ../littlecafe-secrets/certs/jwt-public.pem
chmod 600 ../littlecafe-secrets/certs/jwt-private.pem
```

- Chạy từ repo root; `CERTS_DIR=../../littlecafe-secrets/certs` trong `.env` được
  Compose resolve từ thư mục `deployments/`.
- Không commit `jwt-*.pem`. Nếu file đã tồn tại, giữ nguyên.
- Path trong container: `/run/littlecafe/certs/jwt-{private,public}.pem`.

## 4. Hạ tầng và build

```powershell
make docker-infra-up    # Postgres, Redis, RabbitMQ
make generate
make test
make build-all
```

## 5. Full stack

```powershell
make docker-dev-up      # infra + services + nginx, build local
curl -fsS http://localhost:8080/api/v1/health
```

Kiểm tra CORS:

```bash
curl -i -X OPTIONS http://localhost:8080/api/v1/health \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET"
```

Dừng stack local: `make docker-dev-down`. Chỉ dùng `down -v` khi muốn xóa volume
local (mất DB local).

## 6. mTLS local

Local compose mặc định không bật `GRPC_TLS_*`, nên chỉ cần khi test mTLS:

```bash
sh deployments/scripts/rotate-mtls-certs.sh "$(cd .. && pwd)/littlecafe-secrets/certs"
```

Sau đó thêm `GRPC_TLS_*` theo [`../operations/certificates.md`](../operations/certificates.md).
Cert/key không commit, không đặt trong repo.

## 7. Quyền đọc env staging/production

Cần khi sửa env hoặc chạy tay với secret thật:

```powershell
pwsh scripts/env/init-dev-key.ps1
```

Gửi lead public key `age1...` (không gửi private key); lead thêm vào
`.sops.yaml` và chạy `make env-updatekeys`. Sau khi `git pull`:

```powershell
make env-verify
```

Không tạo `.env.staging`/`.env.production` trên laptop; chỉ decrypt kiểm tra.
Quy trình đầy đủ: [`sops-keys.md`](sops-keys.md) §4.

## 8. Lệnh thường dùng

| Mục đích | Lệnh |
| --- | --- |
| Hạ tầng | `make docker-infra-up` |
| Full stack | `make docker-dev-up` / `make docker-dev-down` |
| Chạy 1 service trên host | `make run svc=commerce` |
| Generate Wire/OpenAPI | `make generate` |
| Xem OpenAPI docs | `make docs-preview` |
| Test | `make test`, `make test-race`, `make test-cover` |
| Coverage gate | `make coverage-check` |
| Lint | `make lint` |
| Gate trước push | `make check` |

## 9. Troubleshooting

- **Trùng port**: dừng stack khác hoặc đổi port trong `deployments/docker-compose.yml`.
- **DB local sai password**: volume cũ giữ password cũ; xóa volume local bằng
  `docker compose -f deployments/docker-compose.yml --env-file .env down -v`.
- **Login lỗi thiếu key**: kiểm tra `jwt-*.pem` trong `CERTS_DIR` (bước 3).
- **mTLS lỗi**: local mặc định tắt; chỉ bật khi đã cấu hình `GRPC_TLS_*`.
- **SOPS `no key could decrypt`**: kiểm tra `SOPS_AGE_KEY_FILE` và quyền đã được
  cấp (bước 7).
- **Docker build lỗi context**: chạy từ repo root; kiểm tra `.dockerignore`.

## 10. Không commit

`.env`, `jwt-*.pem`, cert/key, `keys.txt`/`*.agekey`, ciphertext plaintext.

## 11. Đọc tiếp

| Chủ đề | File |
| --- | --- |
| Config/env, pitfall | [`environments.md`](environments.md) |
| SOPS/age, cập nhật env | [`sops-keys.md`](sops-keys.md) |
| Deploy/rollback | [`../operations/deployment.md`](../operations/deployment.md) |
| Certificate/JWT/mTLS | [`../operations/certificates.md`](../operations/certificates.md) |
| Gate kỹ thuật | [`gates.md`](gates.md) |
| Git/PR/release | [`git-workflow.md`](git-workflow.md) |

## 12. Checklist

- [ ] Cài prerequisites
- [ ] `make env-local` + `make env-local-check`
- [ ] Tạo JWT keys local
- [ ] `make docker-infra-up` + `make generate` + `make test` + `make build-all`
- [ ] `make docker-dev-up` + health/CORS
- [ ] (Khi cần) mTLS local
- [ ] (Khi cần) quyền đọc env staging/production + `make env-verify`
