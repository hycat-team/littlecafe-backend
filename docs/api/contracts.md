# Contract-first

OpenAPI, AsyncAPI/JSON Schema và protobuf trong `contracts/` là nguồn machine-readable chuẩn. Không thêm route trong handler trước contract. HTTP implementation dùng Gin và `oapi-codegen` strict server interface.

URL có major version `/api/v1`. Mutation quan trọng yêu cầu `Idempotency-Key`. Lỗi trả `application/problem+json` theo RFC 9457, thêm `code`, `traceId` và field errors. Cursor pagination là mặc định; offset chỉ dùng cho bảng quản trị.

Protobuf chỉ dành cho read-only cross-service query. Generated code phải commit; CI regenerate và yêu cầu Git diff sạch.

