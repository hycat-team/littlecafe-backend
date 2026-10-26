# Contracts

Repository này là nguồn chuẩn cho machine-readable contracts của LittleCafe. HTTP dùng OpenAPI, sự kiện dùng AsyncAPI/JSON Schema và read-only cross-service dùng protobuf.

Không thêm contract suy đoán. Mọi thay đổi phải truy vết tới requirement, workflow và ADR đã duyệt; generated code được commit và CI kiểm tra drift.

Quy ước API (URL, envelope, pagination, lỗi, auth, idempotency, rate limit, SSE)
nằm tại `littlecafe-docs/api/conventions.md`. OpenAPI dùng các base schema
`SuccessResponse`, `CursorPage`, `AdminPage`, `ProblemDetails` trong
`openapi/components.yaml`.
