# Design Patterns & Conventions

Tài liệu này mô tả các pattern đang có trong `shared` và cách dùng chúng. Khi code mâu thuẫn với tài liệu này, source là bằng chứng triển khai — báo lại để cập nhật tài liệu, không tự chọn một bên.

## 1. Error Pattern (RFC 9457)

### 1.1 Shape canonical

Mọi lỗi HTTP trả về `Content-Type: application/problem+json` theo RFC 9457 với shape của `contracts/openapi/components.yaml`:

| Field | Kiểu | Bắt buộc | Ghi chú |
|---|---|---|---|
| `type` | string (uri-reference) | có | mặc định `about:blank` |
| `title` | string | có | tiêu đề tiếng Việt theo loại lỗi |
| `status` | integer | có | HTTP status |
| `detail` | string | không | message client-facing tiếng Việt |
| `instance` | string | không | request path |
| `code` | string | không | error code nghiệp vụ, xem §1.7 |
| `traceId` | string | không | lấy từ header `X-Trace-ID` do Nginx sinh |
| `errors` | array `{field, message}` | không | lỗi validation theo field |

Hai field debug nằm **ngoài contract**, chỉ xuất hiện khi `SERVER_ENV` không phải production (`config.IsProductionEnv`): `debugMessage` (raw error) và `stackTrace` (đã lọc frame nội bộ). Production không bao giờ trả hai field này.

### 1.2 Package và constructor

`shared/application/constants/apperror` là nguồn duy nhất tạo lỗi:

- `New(status, title, message)` / `NewError(status, title, detail)`: constructor gốc, gắn stack.
- `NewBadRequest`, `NewNotFound`, `NewConflict`, `NewUnauthorized`, `NewForbidden`, `NewTooManyRequest`, `NewRequestEntityTooLarge`, `NewInternalError`, `NewServiceUnavailable`: factory theo HTTP status với title tiếng Việt cố định.
- `NewUpstreamUnavailable(cause)`: 503 an toàn cho lỗi nhà cung cấp; raw error chỉ nằm trong `Cause`, không serialize.
- `static_errors.go`: helper dùng lại (`ErrUnauthorized`, `ErrForbidden`, `ErrInvalidToken`, `ErrInvalidAccessToken`, `ErrUnexpectedSigningToken`, `ErrSearchIndexNotFound`).
- `Wrap(err, fallbackMsg...)` / `From(err)`: bọc lỗi ngoài thành 500; raw text không bao giờ lộ ra client, `Cause` giữ cho log.
- `GenericInternalMessage` và `PanicRecoveryMessage`: hai message client-facing duy nhất cho lỗi không nhận diện và panic.

### 1.3 Luồng xử lý

1. Handler trả `error` qua `presentation.WrapHandler` → `c.Error(err)`.
2. `middleware.GlobalErrorHandler` lấy lỗi cuối, chuyển qua `errorutils.ToAppError` (`apperror.From`/`Wrap`), log Zap, rồi serialize Problem Details.
3. Panic được recover trong cùng middleware; trả 500 với `PanicRecoveryMessage`, stack chỉ ở non-prod.
4. `res.Instance` = request path; `res.TraceID` = `X-Trace-ID`; `Detail` rỗng thì lấy `Message`.

### 1.4 Validation

`shared/pkg/utils/validation`:

- `BindJSON` / `BindQuery` / `BindURI`: bind + `TranslateValidationError` + chuẩn hóa NFC cho string.
- `TranslateValidationError` map `validator.ValidationErrors` sang `errors[]` với message tiếng Việt; `json.UnmarshalTypeError` và body vượt `http.MaxBytesReader` cũng được map (`413`).
- `GetUUIDParam` và helper query trả `NewBadRequest` tiếng Việt.

### 1.5 Error bridge qua gRPC

`contracts/grpcx/errors.go`:

- `ToAppError(err)`: map gRPC status → apperror (NotFound, InvalidArgument/FailedPrecondition, Unauthenticated, PermissionDenied, AlreadyExists/Aborted, ResourceExhausted, Unavailable/DeadlineExceeded); message server (tiếng Việt) được giữ.
- `ToGRPCError(err)`: map ngược apperror → gRPC status để remote client nhận đúng message.
- **Wire contract cho lỗi nghiệp vụ**: `code` đi qua gRPC bằng standard detail `google.rpc.ErrorInfo.Reason`; field errors (dự trữ) đi bằng `google.rpc.BadRequest.FieldViolations`. Khi `code` rỗng, bridge giữ nguyên hành vi cũ (không gửi details).
- **Tương thích**: details là danh sách additive — thêm loại detail mới không phá client cũ; client bỏ qua loại detail không biết. Không dùng message để rẽ nhánh; muốn rẽ nhánh theo lỗi thì so `code`.

### 1.6 Test

Đã có test cho error stack: `apperror/error_test.go`, `apperror/dynamic_errors_test.go`, `errorutils/errorutils_test.go`, `middleware/error_handler_test.go`, `contracts/grpcx/errors_test.go`. Phủ: không leak raw error, shape RFC 9457, gating `debugMessage`/`stackTrace` theo env, panic recovery, guard `Writer.Written()`, map status 2 chiều và round-trip `code` qua gRPC. Khi sửa shape/gating phải cập nhật test tương ứng.

### 1.7 Error code registry

- Format: `<PREFIX>_<NNNN>`, `PREFIX` ∈ `IDENTITY_`, `ORDER_`, `ACTIVITY_`, `WALLET_`, `AI_`, `NOTIFICATION_`, `REPORTING_`; `NNNN` là số tuần tự 4 chữ số cấp theo service, bắt đầu `0001`.
- Code chỉ được thêm cùng OpenAPI operation sở hữu + owner rõ ràng, đăng ký tại `littlecafe-docs/api/error-codes.md` trước khi dùng. Không tái sử dụng/đánh số lại code đã phát hành.
- Lỗi hạ tầng/generic (validation, xác thực, rate limit, internal) **không** gán code.
- Chưa có code nào được cấp; `code` để trống là hợp lệ với contract hiện tại.

## 2. Handler và response

- Handler trả lỗi theo chữ ký `func(c *gin.Context) error`, bọc bằng `presentation.WrapHandler`.
- Response thành công dùng `APIResponse{message, data}` qua `presentation.Success` / `Created` / `Accepted`.
- HTTP contract-first: route khai báo trong `contracts/openapi/components.yaml` và generate strict server interface bằng `oapi-codegen` (`make openapi-gen`); không tự thêm route trong handler trước contract.

## 3. Repository và Unit of Work

- `shared/domain/repositories/generic.go` cung cấp `IGenericRepository[T, ID]` với `GetByID`, `GetByIDs`, `GetAll`, `Create`, `Update`, `Delete`, `BatchUpdate`, `BatchUpdateMap` — ưu tiên prefix `Get`.
- `IUnitOfWork.Execute(ctx, fn)` bọc `gorm.Transaction`; transaction được truyền qua `context.Context` (`db.InjectTx`/`db.GetTx`), không truyền tay.
- Repository lấy DB theo `getDB(ctx)`: dùng transaction trong context nếu có, ngược lại dùng base DB (xem `OutboxRepositoryImpl`).
- `shared/domain/repositories/outbox.go` sở hữu port outbox ở domain layer; implementation GORM nằm ở `shared/infrastructure/messaging` (dependency inversion).
- Mọi truy vấn dùng `WithContext(ctx)` để giữ deadline/cancel.

## 4. Dependency Injection (Wire)

- `internal/di/wire.go` (build tag `wireinject`) khai báo `InitializeApp`; `wire_gen.go` sinh bởi `make wire`/`make wire-all` và **được commit**; CI kiểm tra drift qua `make generate-check`.
- `shared/provider.go` gom `ProviderSet` dùng chung; service thêm provider khi có module thật.
- Thêm/bớt constructor, provider hoặc use case ⇒ chạy lại wire; không có thay đổi graph thì không chạy.

## 5. Worker pattern

- `shared/presentation/worker.BaseWorker`: vòng đời `Start/Stop`, context derive (worker phải dùng `Context()` để graceful shutdown), `running` atomic.
- `OutboxWorker` là worker chuẩn: loop `select` giữa signal `LISTEN/NOTIFY`, poll fallback 2s, gauge 30s và cleanup theo config.
- Worker nghiệp vụ implement `ModuleWorker` (bootstrap) và tự đăng ký consumer khi được khởi tạo.
- Log worker dùng `shared/observability/workerlog`: `worker`, `triggerType`, `runId`, `durationMs`, counters (`totalCount`, `successCount`, `warningCount`, `errorCount`, `skippedCount`, `retryCount`), `result`.

## 6. Rate limiting

- `middleware.RateLimitMiddleware`: Redis fixed-window keyed theo client IP, fail-open khi backend lỗi.
- `middleware.LoginRateLimitMiddleware`: key theo IP + identifier đăng nhập (query/form/Authorization suffix), dùng cho login/OTP.
- Vượt hạn mức trả `429` qua `apperror.NewTooManyRequest` với message tiếng Việt.
- Config: `rate_limit.token_limit`, `rate_limit.replenishment_seconds`.

## 7. SSE và realtime

- `presentation.InitSSE` set header `text/event-stream`, `no-cache`, `keep-alive` và gia hạn write deadline (`DefaultSSEChunkDeadline = 90s`).
- `SSERelay` subscribe Redis Pub/Sub (`sse.RealtimeChannelPattern`) và fan-out vào `ISSEManager` cục bộ (replay buffer + heartbeat theo config). Relay best-effort: lỗi Redis chỉ log + tăng metric, không làm fail request.
- Redis không authoritative; REST fallback là nguồn trạng thái chuẩn.

## 8. Logging và config

- `shared/pkg/logger.Interface` (Zap): console có màu ở dev, JSON ở môi trường khác; file rotation qua lumberjack khi bật.
- `shared/config`: Viper load `configs/config.yaml` + env substitution; `validateConfig` chặn cấu hình sai ngay khi boot (bao gồm `jwt.algorithm=ES256`, outbox batch size, port range).
- `shared/infrastructure/resilience/startup_retry.go` retry kết nối dependency theo lịch `startup.*`.

## 9. Quy ước bổ sung

- JSON DTO client-facing dùng camelCase (`traceId`, `eventType`, `messageId`) — khớp contract đã duyệt.
- Client-facing message tiếng Việt; identifier/package/field code giữ tiếng Anh.
- Không tạo entity/route/RPC/event/fixture giả để minh họa.
