# Frontend Integration Guide — API message localization

## 1. Summary

- Feature/spec: `specs/duck-001-api-localization/`
- Requirement/workflow: cross-cutting decision DEC-006 / ADR-001 (littlecafe-docs)
- Handoff status: READY_FOR_FRONTEND
- Canonical contracts: `contracts/openapi/components.yaml` (AcceptLanguage, ContentLanguage), `contracts/asyncapi/asyncapi.yaml` (EventEnvelope.locale)
- Counts: new endpoints 0 · changed 0 · removed 0 · events 1 · DTOs 0

## 2. Authentication and behavior

- Authentication/authorization: unchanged.
- Idempotency: unchanged.
- Pagination: unchanged.
- Realtime/SSE behavior and REST fallback: unchanged.

## 3. Endpoints

| Method | Path | Status (new/changed/removed) | Before → After (changed only) |
|---|---|---|---|
|  |  |  |  |

Không có endpoint mới/thay đổi/xóa. Mọi endpoint hiện có nhận thêm header tùy chọn
`Accept-Language` và trả `Content-Language` + `Vary: Accept-Language`.

## 4. Events

| Topic / routing key | Payload summary | Status (new/changed/removed) | Before → After |
|---|---|---|---|
| mọi event dùng `EventEnvelope` | thêm field tùy chọn `locale` (`vi`/`en`) | changed (additive) | không có `locale` → `locale` optional; thiếu thì consumer dùng `vi` |

## 5. Client-facing DTOs

| DTO | Field | Type | Status | Before → After |
|---|---|---|---|---|
|  |  |  |  |  |

Không có DTO HTTP nào đổi field. `ProblemDetails` và `APIResponse` giữ nguyên shape;
chỉ nội dung chuỗi được localize theo `Accept-Language`.

## 6. Enums

| Enum | Value | Status (new/changed/removed) |
|---|---|---|
| locale | `vi` | new |
| locale | `en` | new |

`Accept-Language` được negotiate; giá trị không hỗ trợ fallback về `vi`.

## 7. Errors and compatibility

- Problem Details/error codes: shape RFC 9457 không đổi. `title`, `detail`,
  `errors[].message` được render theo locale (`vi` mặc định, `en` hỗ trợ). Error
  `code` không đổi.
- Migration/compatibility: client cũ không gửi `Accept-Language` nhận tiếng Việt
  như trước (tương thích ngược). Không có breaking change.
- Test evidence:
  - `go test ./...` trong `shared` (i18n, apperror, validation, middleware, config)
  - `go test ./grpcx/...` trong `contracts`
  - `make openapi-gen` không drift
  - `make build-all`

## 8. Sync checklist (fill counts before closing the feature)

- [x] Every JSON field the feature's DTOs emit appears above (count: 0 changed)
- [x] Every value of every enum compared against its source constant block, in
      both directions (count: 2 locales = `vi`, `en`, matching
      `i18n.SupportedLocales`)
- [x] Every route compared after normalizing path params (`:param` vs `{param}`)
      (count: 0 changed)
- [x] Every HTTP status claimed above is one the handler can actually return
      (count: 0 changed)
