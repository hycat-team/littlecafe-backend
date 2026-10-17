# Bảo mật

Tài liệu này mô tả các cơ chế bảo mật đang có trong code và các quy tắc đã chốt. Quy tắc bắt buộc nằm trong constitution mục V; khi code chưa đủ, ghi rõ "chưa triển khai".

## 1. Xác thực người dùng

### 1.1 Email/password

- Hash bằng **Argon2id** (`services/identity/internal/modules/identity/infrastructure/security/passwordutils/argon2id.go`): salt 16 byte ngẫu nhiên, `time=3`, `memory=64 MiB`, `parallelism=2`, key 32 byte; verify bằng so sánh constant-time.
- Password rỗng bị từ chối ngay từ `Hash`.

### 1.2 Google OAuth/OIDC

- Adapter: `services/identity/internal/modules/identity/infrastructure/oauth/google_provider.go`, cấu hình `google.*` trong `configs/identity.yaml`.
- Redirect URI chuẩn: `https://api.littlecafe.kids/api/v1/identity/oauth/google/callback`; state TTL mặc định 10 phút.
- Sau callback, backend đổi code lấy identity; redirect về frontend thuộc feature Identity và chưa chốt env key (hiện chỉ có `GOOGLE_OAUTH_CALLBACK_URL` là callback backend).

### 1.3 Access token ES256

- Ký ES256 ở Identity (`services/identity/internal/modules/identity/infrastructure/security/jwtsigner`), verify ở `shared/pkg/utils/jwtutils` (`LoadECDSAPublicKey`, `ValidateTokenES256`). **Chỉ Identity giữ private key** (`signing.private_key_file` trong `configs/identity.yaml`, bắt buộc ở production theo `Config.Validate`); service khác chỉ nạp public key (`jwt.public_key_file`, bắt buộc ở production theo `validateConfig`).
- Claims: `sub` (user UUID), `email`, `role_slug`, `type`, `iss`, `aud`, `exp`, `iat`, `jti`.
- `ValidateTokenES256` chặn sai thuật toán (phải đúng `ES256`) và sai token type; lỗi trả `apperror.ErrInvalidAccessToken`/`ErrUnexpectedSigningToken`.

### 1.4 Refresh token

- Sinh 32 byte ngẫu nhiên, encode Base64URL; **chỉ lưu SHA-256 digest** (`services/identity/internal/modules/identity/infrastructure/security/tokenutils/refresh.go`).
- Rotation, revoke và reuse detection là yêu cầu bắt buộc của feature Identity (constitution V); chưa có implementation trong bootstrap.
- Thu hồi phiên dùng `shared/infrastructure/security/redis_token_blacklist_service.go` (Redis dùng chung): blacklist theo token, theo user, có prefix tuỳ biến; `CheckBlacklist` kiểm tra cả hai.

## 2. Service-to-service (gRPC)

- **mTLS** qua `contracts/grpcx/mtls.go`: TLS tối thiểu 1.3; client verify CA + gửi client cert; server `RequireAndVerifyClientCert`.
- Env: `GRPC_TLS_CA_FILE`, `GRPC_TLS_CERT_FILE`, `GRPC_TLS_KEY_FILE`; `GRPC_MTLS_REQUIRED=true` bắt buộc có đủ 3 path, thiếu → fail ngay khi khởi động (fail closed).
- Certificate do private CA cấp, nằm ngoài Git, mount read-only tại `/run/littlecafe/certs`. Script `deployments/scripts/rotate-mtls-certs.sh` tạo CA/server/client cert; leaf 90 ngày, CA 10 năm (`LITTLECAFE_CERT_DAYS`, `LITTLECAFE_CA_DAYS`). Chi tiết vận hành: `docs/operations/certificates.md`.
- **User JWT đi kèm khi service hành động thay người dùng**: `contracts/grpcx/metadata.go` truyền metadata `x-littlecafe-jwt`; mTLS xác thực service, metadata xác thực người dùng.
- Lỗi gRPC được bridge giữ nguyên status, message client-facing và `code` (qua `google.rpc.ErrorInfo.Reason`); locale lan qua metadata `x-littlecafe-locale` và `ErrorInfo.Metadata` (`messageKey`, `locale`); field errors dự trữ cho `google.rpc.BadRequest.FieldViolations`. Details là additive, client bỏ qua loại không biết.

## 3. Phân quyền (RBAC)

- Vai trò hệ thống chỉ có `ADMIN` và `USER` (`shared/domain/constants/roleslug`); không thêm RoleSlug khác.
- Kết hợp **ownership**, **Guardian-Child** (mọi dữ liệu trẻ em phải qua `GuardianLink`) và **assignment** cho vận hành.
- Chi tiết actor/quyền theo `littlecafe-docs/business/actors-and-permissions.md` và workflow tương ứng.
- Helper context: `contextutils.GetUserId`, `GetRoleSlug`, `GetAccessToken` trả `ErrUnauthorized` khi thiếu.

## 4. Dữ liệu nhạy cảm

- **AES-GCM** (`shared/pkg/utils/cryptoutils/aesgcm.go`): nonce ngẫu nhiên, hỗ trợ associated data. Dùng cho encrypted snapshot của withdrawal với withdrawal ID làm associated data.
- Không lưu bank profile tái sử dụng; mỗi withdrawal lưu snapshot mã hoá (theo `littlecafe-docs/integrations/payos.md`).
- PII không xuất hiện trong log, event hoặc fixture ngoài mức tối thiểu.
- Secret và certificate nằm ngoài Git; `.env` thật không commit.

## 5. API surface

- Lỗi trả `application/problem+json` theo RFC 9457 kèm `code` (khi có), `traceId`, `errors` — xem `design-pattern.md` §1.
- `traceId` lấy từ `X-Trace-ID` do Nginx sinh (`$request_id`). Gọi trực tiếp service không qua Nginx thì header trống và `traceId` rỗng.
- Mutation quan trọng yêu cầu `Idempotency-Key` (xem `littlecafe-docs/api/conventions.md`).
- Rate limit: shield theo IP dùng Redis token bucket (`RateLimitMiddleware`), quota nghiệp vụ và login/OTP dùng Redis fixed-window (`IRateLimiter`); cả hai fail-open khi Redis lỗi, Nginx `limit_req` là shield ở biên.

## 6. Chưa triển khai

- Rotation/revoke/reuse detection cho refresh token (thuộc feature Identity).
- Middleware xác thực/authorization thực tế cho route nghiệp vụ (chưa có route nghiệp vụ).
- Test bảo mật: `shared` có test cho AES-GCM (`cryptoutils`); Identity có test cho Argon2id, refresh token và ES256 signer (`infrastructure/security/{passwordutils,tokenutils,jwtsigner}`); `shared/pkg/utils/jwtutils` chưa có test.
