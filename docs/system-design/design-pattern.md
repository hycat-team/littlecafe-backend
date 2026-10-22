# Design Patterns & Conventions

Tài liệu này mô tả các pattern đang có trong `shared` và cách dùng chúng. Khi code mâu thuẫn với tài liệu này, source là bằng chứng triển khai — báo lại để cập nhật tài liệu, không tự chọn một bên.

> Đọc section liên quan khi implement; không cần đọc toàn bộ file. Ranh giới
> service, contract map và gRPC transport: [`architecture.md`](architecture.md)
> §7–8.

## 1. Error Pattern (RFC 9457)

### 1.1 Shape canonical

Mọi lỗi HTTP trả về `Content-Type: application/problem+json` theo RFC 9457 với shape của `contracts/openapi/components.yaml`:

| Field | Kiểu | Bắt buộc | Ghi chú |
|---|---|---|---|
| `type` | string (uri-reference) | có | mặc định `about:blank` |
| `title` | string | có | tiêu đề theo loại lỗi, localize theo `Accept-Language` (mặc định `vi`) |
| `status` | integer | có | HTTP status |
| `detail` | string | không | message client-facing, localize theo `Accept-Language` |
| `instance` | string | không | request path |
| `code` | string | không | error code nghiệp vụ, xem §1.7 |
| `traceId` | string | không | lấy từ header `X-Trace-ID` do Nginx sinh |
| `errors` | array `{field, message}` | không | lỗi validation theo field |

Hai field debug nằm **ngoài contract**, chỉ xuất hiện khi `SERVER_ENV` không phải production (`config.IsProductionEnv`): `debugMessage` (raw error) và `stackTrace` (đã lọc frame nội bộ). Production không bao giờ trả hai field này.

### 1.2 Package và constructor

`shared/application/constants/apperror` là nguồn duy nhất tạo lỗi:

- `New(status, titleKey, messageKey, args...)` / `NewError(status, titleKey, detailKey, args...)`: constructor gốc, nhận **message key** trong `shared/pkg/i18n` và gắn stack. Key không có trong catalog được giữ nguyên (dùng cho message passthrough từ service khác).
- `NewBadRequest`, `NewNotFound`, `NewConflict`, `NewUnauthorized`, `NewForbidden`, `NewTooManyRequest`, `NewRequestEntityTooLarge`, `NewInternalError`, `NewServiceUnavailable`: factory theo HTTP status; title lấy từ catalog (`error.title.*`).
- `NewUpstreamUnavailable(cause)`: 503 an toàn cho lỗi nhà cung cấp; raw error chỉ nằm trong `Cause`, không serialize.
- `static_errors.go`: helper dùng lại (`ErrUnauthorized`, `ErrForbidden`, `ErrInvalidToken`, `ErrInvalidAccessToken`, `ErrUnexpectedSigningToken`, `ErrSearchIndexNotFound`).
- `Wrap(err, fallbackKey...)` / `From(err)`: bọc lỗi ngoài thành 500; raw text không bao giờ lộ ra client, `Cause` giữ cho log.
- `GenericInternalMessage` và `PanicRecoveryMessage`: hai **key** message client-facing duy nhất cho lỗi không nhận diện và panic.
- `Error.Localize(loc)`: trả bản copy đã render `title/detail/message` theo locale; middleware dùng hàm này trước khi serialize. `Error.Message`/`Title`/`Detail` luôn chứa bản render default-locale để log và các caller không có request context vẫn hoạt động.

### 1.3 Luồng xử lý

1. `middleware.LocaleMiddleware` (đăng ký sau `GlobalErrorHandler`) parse `Accept-Language`, chọn locale trong `vi`/`en`, lưu vào `context.Context` của request và set `Content-Language`.
2. Handler trả `error` qua `presentation.WrapHandler` → `c.Error(err)`.
3. `middleware.GlobalErrorHandler` lấy lỗi cuối, chuyển qua `errorutils.ToAppError` (`apperror.From`/`Wrap`), gọi `Localize` theo locale của request, log Zap, rồi serialize Problem Details.
4. Panic được recover trong cùng middleware; trả 500 với `PanicRecoveryMessage`, stack chỉ ở non-prod.
5. `res.Instance` = request path; `res.TraceID` = `X-Trace-ID`; `Detail` rỗng thì lấy `Message`.

Router phải bật `gin.Engine.ContextWithFallback = true` để `i18n.FromContext(c)` đọc được locale đã lưu trên request context.

### 1.4 Validation

`shared/pkg/utils/validation`:

- `BindJSON` / `BindQuery` / `BindURI`: bind + `TranslateValidationError(ctx, err)` + chuẩn hóa NFC cho string.
- `TranslateValidationError` map `validator.ValidationErrors` sang `errors[]`; message resolve từ catalog (`validation.*`) theo locale của request; field label resolve từ `field.<jsonName>` (fallback humanize). `json.UnmarshalTypeError` và body vượt `http.MaxBytesReader` cũng được map (`413`).
- `GetUUIDParam` và helper query trả `NewBadRequest` với key `validation.param.*`; middleware localize khi serialize.

### 1.5 Error bridge qua gRPC

`contracts/grpcx/errors.go`:

- `ToAppError(err)`: map gRPC status → apperror (NotFound, InvalidArgument/FailedPrecondition, Unauthenticated, PermissionDenied, AlreadyExists/Aborted, ResourceExhausted, Unavailable/DeadlineExceeded); message server (đã localize) được giữ.
- `ToGRPCError(err)` / `ToGRPCErrorWithContext(ctx, err)`: map ngược apperror → gRPC status, render theo locale trong `ctx` (bản không có ctx dùng default locale).
- **Wire contract cho lỗi nghiệp vụ**: `code` đi qua gRPC bằng standard detail `google.rpc.ErrorInfo.Reason`; khi có `code`, `ErrorInfo.Metadata` mang thêm `messageKey` và `locale` để service nhận có thể re-render. Field errors (dự trữ) đi bằng `google.rpc.BadRequest.FieldViolations`. Khi `code` rỗng, bridge giữ nguyên hành vi cũ (không gửi details).
- **Locale propagation**: `grpcx.ContextWithLocale` ghi `x-littlecafe-locale`; server options mặc định gắn unary/stream interceptor đọc metadata vào `context.Context`; client interceptor trong `DialService` tự forward locale từ ctx.
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
- Response thành công dùng `APIResponse{message, data}` qua `presentation.Success` / `Created` / `Accepted`; dùng `LocalizedSuccess` / `LocalizedCreated` / `LocalizedAccepted` (hoặc `presentation.Message`) với message key để resolve theo locale.
- HTTP contract-first: route khai báo trong `contracts/openapi/components.yaml` và generate strict server interface bằng `oapi-codegen` (`make openapi-gen`); không tự thêm route trong handler trước contract.
- **Auth middleware (forward-convention; code do feature Identity implement)**: `Handle()` bắt buộc JWT hợp lệ (401 nếu thiếu/sai), `OptionalHandle()` parse nếu có nhưng không fail khi thiếu. Cả hai set `CtxUserId`/`CtxRoleSlug`/`CtxAccessToken` (`shared/pkg/utils/contextutils`). Phân quyền là middleware riêng `RolesAuthorize(roleslug.ADMIN, roleslug.USER)`; RBAC không thay thế ownership hay Guardian-Child.
- Message success/error khai báo constant **message key** ở đầu file handler (xem §10.3 và §16).

## 3. Repository và Unit of Work

- `shared/domain/repositories/generic.go` cung cấp `IGenericRepository[T, ID]` với `GetByID`, `GetByIDs`, `GetAll`, `Create`, `Update`, `Delete`, `BatchUpdate`, `BatchUpdateMap` — ưu tiên prefix `Get`.
- `IUnitOfWork.Execute(ctx, fn)` bọc `gorm.Transaction`; transaction được truyền qua `context.Context` (`db.InjectTx`/`db.GetTx`), không truyền tay.
- Repository lấy DB theo `getDB(ctx)`: dùng transaction trong context nếu có, ngược lại dùng base DB (xem `OutboxRepositoryImpl`).
- `shared/domain/repositories/outbox.go` sở hữu port outbox ở domain layer; implementation GORM nằm ở `shared/infrastructure/messaging` (dependency inversion).
- Mọi truy vấn dùng `WithContext(ctx)` để giữ deadline/cancel.

### 3.1 Generic repository implementation

`shared/infrastructure/repositories/generic_impl.go` cung cấp `GenericRepository[T, ID]` implement đủ `IGenericRepository` (gồm `GetByIDs`, `BatchUpdate`, `BatchUpdateMap`):

- Preload cố định lúc khởi tạo: `NewGenericRepository[T, ID](db, "Relation")` để mọi method kế thừa load cùng quan hệ.
- Transaction-aware: `GetDB(ctx)` trả transaction trong context (do `IUnitOfWork.Execute` bind qua `db.InjectTx`) nếu có, ngược lại base DB. Custom query MUST dùng `GetDB(ctx)`/`GetQueryWithPreload(ctx)`.
- `GetAll`/`GetByIDs` trả slice rỗng (không `nil`) khi không có dữ liệu.

Repository domain-specific embed generic và thêm method riêng; interface đặt ở `domain/repositories/`, implementation ở `infrastructure/persistence/`.

### 3.2 Unit of Work trong use case

```go
if err := uc.uow.Execute(ctx, func(txCtx context.Context) error {
    if err := uc.orderRepo.Create(txCtx, order); err != nil {
        return err
    }
    return uc.outboxRepo.Create(txCtx, record)
}); err != nil {
    return err
}
```

Mọi thao tác trong `fn` chia sẻ một transaction; lỗi bất kỳ → rollback toàn bộ. Không truyền `*gorm.DB` bằng tay.

## 4. Dependency Injection (Wire)

- `internal/di/wire.go` (build tag `wireinject`) khai báo `InitializeApp`; `wire_gen.go` sinh bởi `make wire`/`make wire-all` và **được commit**; CI kiểm tra drift qua `make generate-check`.
- `shared/provider.go` gom `ProviderSet` dùng chung; service thêm provider khi có module thật.
- Thêm/bớt constructor, provider hoặc use case ⇒ chạy lại wire; không có thay đổi graph thì không chạy.

## 5. Worker pattern

- `shared/presentation/worker.BaseWorker`: vòng đời `Start/Stop`, context derive (worker phải dùng `Context()` để graceful shutdown), `running` atomic.
- `OutboxWorker` là worker chuẩn: loop `select` giữa signal `LISTEN/NOTIFY`, poll fallback 2s, gauge 30s và cleanup theo config.
- Worker nghiệp vụ implement `ModuleWorker` (bootstrap) và tự đăng ký consumer khi được khởi tạo.
- Log worker dùng `shared/observability/workerlog`: `worker`, `triggerType`, `runId`, `durationMs`, counters (`totalCount`, `successCount`, `warningCount`, `errorCount`, `skippedCount`, `retryCount`), `result`.

### 5.1 Quy tắc

- Worker embed `*worker.BaseWorker`; `Start(ctx)` gọi `BaseWorker.Start(ctx)` trước, `Stop()` gọi `BaseWorker.Stop()` sau.
- Goroutine nền MUST dùng `w.Context()` (đã derive) để graceful shutdown, không dùng `ctx` thô.
- Worker là wrapper mỏng: delegate business logic cho use case. Consumer đăng ký qua `messaging.RegisterTypedHandler(queue, handler, opts)`; `event.RetryableError` → requeue, `event.PermanentError` → DLQ, lỗi không phân loại → DLQ (kèm cảnh báo log).
- `bootstrap.Workers` gom `*OutboxWorker` + danh sách `ModuleWorker`; field worker dùng concrete type, không dùng interface.

## 6. Rate limiting

Hai cơ chế Redis, chọn theo ngữ nghĩa (không gộp "cho nhất quán"):

- **Shield theo IP (token bucket / GCRA)** — `middleware.RateLimitMiddleware` + `caching.RedisTokenBucketLimiter` (`ITokenBucketLimiter`): throttling mượt, key theo `c.ClientIP()`; `token_limit` = burst, `tokens_per_period`/`replenishment_seconds` = tốc độ refill. Atomic Lua + Redis `TIME` nên chính xác giữa các instance. Fail-open khi Redis lỗi; Nginx `limit_req` vẫn là shield thô ở biên.
- **Quota nghiệp vụ (fixed window)** — `caching.RedisRateLimiter` (`IRateLimiter.Allow(key, limit, windowSeconds)`): "N lần / chu kỳ" chính xác cho một subject, dùng ở use case và cho login/OTP (`LoginRateLimitMiddleware` của Identity).
- Vượt hạn mức trả `429` qua `apperror.NewTooManyRequest` với message key (`error.rate_limit.exceeded`, `error.login.rate_limit_exceeded`).
- Config `rate_limit.*` được validate khi boot (`token_limit`, `tokens_per_period`, `replenishment_seconds` phải > 0).
- `ClientIP()` phải resolve đúng client thật sau proxy/LB (Gin trusted proxy); nếu không mọi user chung một bucket.

## 7. SSE và realtime

- `presentation.InitSSE` set header `text/event-stream`, `no-cache`, `keep-alive` và gia hạn write deadline (`DefaultSSEChunkDeadline = 90s`).
- `SSERelay` subscribe Redis Pub/Sub (`sse.RealtimeChannelPattern`) và fan-out vào `ISSEManager` cục bộ (replay buffer + heartbeat theo config). Relay best-effort: lỗi Redis chỉ log + tăng metric, không làm fail request.
- Redis không authoritative; REST fallback là nguồn trạng thái chuẩn.

## 8. Logging và config

- `shared/pkg/logger.Interface` (Zap): console có màu ở dev, JSON ở môi trường khác; file rotation qua lumberjack khi bật.
- `shared/config`: Viper load `configs/base.yaml` merge overlay `configs/<service>.yaml` + env substitution; `validateConfig` chặn cấu hình sai ngay khi boot (bao gồm `jwt.algorithm=ES256`, outbox batch size, port range). Service có section riêng dùng package config của mình và embed `shared/config.Config`.
- `shared/infrastructure/resilience/startup_retry.go` retry kết nối dependency theo lịch `startup.*`.

## 9. Quy ước bổ sung

- JSON DTO client-facing dùng camelCase (`traceId`, `eventType`, `messageId`) — khớp contract đã duyệt.
- Client-facing message localize (`vi` mặc định, `en` hỗ trợ); identifier/package/field code giữ tiếng Anh. Xem §16.
- Không tạo entity/route/RPC/event/fixture giả để minh họa.

## 10. Quy ước constants

Áp dụng cho hằng số enum/domain (status, type, code, key, reason...) trong `shared/` và trong service. Không áp dụng cho message key (xem §16), literal của protocol bên thứ ba, và giá trị đã phát hành ra contract/DB.

### 10.1 Giá trị string

- Hằng số string MUST dùng `UPPER_SNAKE_CASE`: `"NEEDS_REVIEW"`, `"ACCESS_TOKEN"`, `"WALLET_0007"`.
- Toàn bộ enum nội bộ đã migrate sang UPPER_SNAKE: `roleslug` (`"ADMIN"`, `"USER"`), `jwttype` (`"ACCESS_TOKEN"`, `"REFRESH_TOKEN"`...), outbox status (`"PENDING"`, `"PUBLISHED"`...), context key (`"ACCESS_TOKEN"`, `"USER_ID"`...), workerlog trigger/result (`"CRON"`, `"SUCCESS"`...).
- Ngoại lệ không đổi value:
  - Literal của protocol/thư viện: AMQP queue type (`"classic"`, `"quorum"`), exchange type (`"topic"`), URL literal (`"https://api.cloudinary.com"`, `"amqps"`, `"https"`), Cloudinary resource type (`"image"`), gRPC metadata key (`"x-littlecafe-jwt"` — gRPC yêu cầu lowercase).
  - Tên exchange/queue/routing key theo quy ước messaging (`littlecafe.events`, `<queue>.dlq`) và Redis key prefix theo namespace (`littlecafe:...`, `rl:...`, `blacklist:...`).

### 10.2 Giá trị số

- Hằng số dạng số MUST NOT nhận giá trị `0`. `0` giữ nghĩa zero-value/unset của type; hằng số enum bắt đầu từ `1`.
- Không dùng `0` làm phần tử đầu của `iota`; nếu dùng `iota`, viết `iota + 1`.
- Type vẫn có thể nhận `0` ở dữ liệu runtime (zero-value hợp lệ), nhưng `0` MUST NOT được định nghĩa thành một hằng số enum.

### 10.3 Cách dùng

- Luôn tham chiếu symbol, không hardcode value tại call site (kể cả test), trừ khi test đang kiểm tra chính giá trị wire/contract.
- Hằng số dùng chung nhiều service đặt tại `shared/application/constants/<name>` hoặc `shared/domain/constants/<name>`; hằng số chỉ một service dùng đặt trong service đó (theo rule shared-code ownership).
- Message success/error của handler khai báo thành constant ở đầu file handler, không hardcode inline.

## 11. Utility Packages

`shared/pkg/utils` chỉ chứa building block dùng chung (theo shared-code ownership):

| Package | Dùng cho |
|---|---|
| `contextutils` | `GetUserId`/`GetUserIdOptional`/`GetRoleSlug`/`GetAccessToken` đọc giá trị auth middleware set trên gin context; key ở `key.go` |
| `validation` | `BindJSON`/`BindQuery`/`BindURI`, `GetUUIDParam`/`GetStringParam`/`GetInt64Param`; validator `username`, `neqfield`, `password_complexity`, `nfcmax`; message localize theo request locale |
| `errorutils` | `ToAppError`, `WrapError`, `FilterStackTraceArray`, `PrimaryStackFrame` |
| `numberutils` | `FormatBytes` |
| `sliceutils` | `IndexBy`, `UniqueUUIDs`, `UniqueAndSortStrings`, ... |
| `jwtutils` | `ValidateTokenES256`, load EC public key (verify dùng chung; signer ES256 nằm ở Identity `infrastructure/security/jwtsigner`) |
| `cryptoutils` | `Encrypt`/`Decrypt` AES-GCM |
| `timeutils` | `BusinessLocation`, `FormatAPI`, `GetNow` (Asia/Ho_Chi_Minh) |

Guideline: handler MUST dùng `validation.Bind*` thay `c.ShouldBind*` để có message localize nhất quán.

## 12. Pagination

Cursor là mặc định cho client-facing; offset chỉ cho bảng admin.

- `shared/application/dto/pagination.go`: `CursorPage[T]{Items, NextCursor, HasMore}`, `EncodeCursor`/`DecodeCursor` (base64 URL, opaque); `AdminPage[T]{Items, Page, Limit, TotalItems}`.
- Repository admin helper: `repositories.ApplyAdminPagination(db, page, limit)` (offset; mặc định page=1, limit=20).
- Recipe cursor khi implement feature: filter có `Cursor` + `Limit`; query `ORDER BY <sort_key> DESC, id DESC` với `WHERE (<sort_key>, id) < (?, ?)` từ cursor đã decode; lấy `limit+1` để suy ra `HasMore`; trả `NextCursor` từ phần tử cuối; handler bind `limit`/`cursor`; cursor không trả `totalItems`.
- Không dùng offset cho client-facing (random access, chậm ở trang sâu).

## 13. Module structure, mapper & helper

Trách nhiệm layer (layout ở `architecture.md §2`):

- `presentation/{handler,router,grpc,worker}`: nhận request, gọi use case, trả qua `presentation.WrapHandler`; message constant ở đầu file handler.
- `application/{usecase,ports,dto,mapper,errors}`: business logic/orchestration; `ports` là interface ra ngoài; `mapper` chuyển entity ↔ DTO; `errors` compose `apperror` cho lỗi domain.
- `domain/{entities,repositories,constants}`: entity GORM, repository interface (+ Filter/ListResult), enum/constant; không phụ thuộc layer ngoài.
- `infrastructure/persistence`: implement repository.

Quy ước:

- Mapper đặt trong `application/mapper`, không để trong file use case; có thể tách theo nhóm nghiệp vụ.
- Helper chỉ phục vụ một use case đặt cùng package, tên file hậu tố `_helper.go`; không tạo helper nếu chỉ bọc một biểu thức đơn giản.
- Contract liên module đặt ở `presentation/contract` (xem §15), không dùng cho HTTP DTO.

## 14. DTO convention (contract-first)

- OpenAPI, AsyncAPI/JSON Schema và protobuf trong `contracts/` là nguồn
  machine-readable chuẩn. Không thêm route trong handler trước contract.
- HTTP Req/Res là **type sinh từ OpenAPI** (`contracts/openapi/generated.gen.go` qua oapi-codegen strict server). Không viết tay Req/Res HTTP trong `application/dto` — contract là nguồn chuẩn, sửa contract rồi generate.
- URL có major version `/api/v1`; mutation quan trọng yêu cầu `Idempotency-Key`.
- Cursor pagination là mặc định; offset chỉ dùng cho bảng quản trị.
- Protobuf chỉ dùng cho read-only cross-service query. Generated code phải
  commit; CI regenerate và yêu cầu Git diff sạch.
- `application/dto` giữ các shape không sinh từ OpenAPI: contract payload liên module, event payload, read-model/analytics, alias pagination.
- Enum/constant không đặt trong `dto`; để ở `domain/constants`.
- File nhóm theo feature (không tách `_req.go`/`_res.go`, không file một-struct).
- Naming: `XxxContractDTO` (contract payload), `XxxEventPayload` (event), `XxxStatsDTO`/`XxxAnalyticsDTO` (read model), `XxxInput` (input phi HTTP). List contract trả slice rỗng, không `nil`.
- Không tái dùng một struct cho cả API response và contract payload.

## 15. Cross-module contracts & gRPC transport (A′)

Trong một service, module giao tiếp qua `presentation/contract/`:

- Contract chỉ expose DTO (`application/dto`), không expose entity.
- Implementation contract delegate cho use case qua interface.
- Module MUST NOT import `domain/entities`, `domain/repositories`, `infrastructure/persistence` của module khác.
- Ngoại lệ: được import `domain/constants` của module khác (giá trị enum thuần); luôn tham chiếu symbol, không hardcode value.

Qua biên service (gRPC): wire types là protobuf trong `contracts/proto`; mapping wire ⇄ contract-DTO do service sở hữu đặt tại:

```text
services/<service>/contract/<owner>/transport/mapper.go
```

- Package `transport` nằm ngoài `internal/` để service khác import được; được tham chiếu `domain/constants` của owner khi dựng DTO.
- Generated client (`contracts/<owner>/.../client`) là adapter mỏng: marshal, gọi stub, map lỗi transport, delegate mapping cho `transport`; MUST NOT import domain enum/constant của owner.
- Cùng package `transport` dùng lại cho gRPC server trong service → một nguồn mapping duy nhất.
- Lỗi qua biên service đi qua bridge `contracts/grpcx/errors.go` (xem §1.5).

## 16. Đa ngôn ngữ (i18n)

Phạm vi: mọi nội dung **trả ra client** — response HTTP (Problem Details, validation, success message), email/SMS và nội dung notification. Log nội bộ vẫn tiếng Anh; khi log lỗi API thì log đúng chuỗi đã trả cho client.

### 16.1 Locale

- Hỗ trợ `vi` (mặc định) và `en`; danh sách chuẩn ở `shared/pkg/i18n` (`SupportedLocales`).
- Nguồn locale: header `Accept-Language`; thiếu hoặc không hỗ trợ → default locale.
- Response set `Content-Language` và `Vary: Accept-Language`.
- Negotiation dùng `golang.org/x/text/language` (không thêm dependency mới).

### 16.2 Catalog và message key

- Catalog: `shared/pkg/i18n`, bundle JSON embed theo namespace `locales/<locale>/<domain>.json` (`error`, `validation`, `field`, `success`); loader merge tất cả file của locale, từ chối key trùng giữa các file; `Validate()` kiểm tra key-set khớp giữa các locale.
- Key đặt tên theo namespace: `error.title.*`, `error.*`, `validation.*`, `field.*`, `success.*`, `notification.*`.
- Key dùng chung khai báo trong `shared/pkg/i18n/keys.go`; message riêng của service đăng ký lúc boot:

```go
//go:embed locales/*.json
var bundleFS embed.FS

func registerMessages() {
    i18n.Default().Register(i18n.VI, viMessages)
    i18n.Default().Register(i18n.EN, enMessages)
}
```

- Bundle của service nằm trong service đó (shared-code ownership); test phải kiểm tra key-set parity và sự tồn tại của mọi key service dùng.

### 16.3 Request flow

- `middleware.LocaleMiddleware` chạy sau `GlobalErrorHandler`, lưu locale vào `context.Context` (`i18n.WithLocale`) và `Content-Language`.
- Error: `apperror` mang `TitleKey`/`MessageKey`/`MessageArgs`; `GlobalErrorHandler` gọi `Localize(locale)` trước khi serialize.
- Success: handler dùng `presentation.LocalizedSuccess/Created/Accepted` (message key).
- Validation: `TranslateValidationError(ctx, err)` resolve message theo locale của request; item trong `errors[]` đã localize.
- Đọc locale trong handler: `i18n.FromContext(c)` (cần `gin.Engine.ContextWithFallback = true`) hoặc `i18n.FromContext(c.Request.Context())`.

### 16.4 Cross-service và event

- gRPC: locale đi qua metadata `x-littlecafe-locale`; `grpcx.ServerOptionsFromEnv` gắn interceptor đọc locale vào ctx, `DialService` tự forward locale. `ToGRPCErrorWithContext` render theo locale và (khi có `code`) gửi `messageKey`/`locale` trong `ErrorInfo.Metadata`.
- Event: `EventEnvelope.locale` (AsyncAPI, additive) mang locale của user; consumer gọi `i18n.WithLocale(ctx, locale)` trước khi render notification. Không có locale → default `vi`.
- Notification/Reporting không đọc DB service khác; locale lấy từ event hoặc preference do owner cung cấp.

### 16.5 Quy ước

- Không hardcode câu client-facing trong handler/use case; luôn dùng key.
- Không đặt nội dung client-facing trong log message; log message giữ tiếng Anh.
- Thêm key mới phải thêm đủ `vi` và `en` trong cùng thay đổi; key thiếu bản dịch làm `Catalog.Validate()` fail khi boot/test.
- Feature đổi hành vi message/header phải cập nhật
  `specs/<owner>-<NNN>-<slug>/frontend-guide.md`.
