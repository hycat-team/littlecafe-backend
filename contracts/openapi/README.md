# OpenAPI

- Server production chuẩn: `https://api.littlecafe.kids/api/v1`.
- Dùng `oapi-codegen` strict server interface với Gin.
- Chạy `make openapi-gen`; file `generated.gen.go` phải được commit và CI kiểm tra drift.
- Mọi mutation quan trọng khai báo header `Idempotency-Key`.
- Lỗi dùng RFC 9457 Problem Details; `type` mặc định `about:blank`, `code` chỉ
  dùng cho lỗi nghiệp vụ đã đăng ký tại `littlecafe-docs/api/error-codes.md`.
- Response thành công dùng `SuccessResponse` (`message`, `data`); `data` có thể
  `null`, trừ `204` không body.
- Pagination dùng `CursorPage` (`items`, `nextCursor`, `hasMore`) cho client;
  `AdminPage` chỉ cho bảng quản trị.
- OpenAPI 3.0.3 không có generic: base schema + `allOf` cho từng resource.
- Auth khai báo `bearerAuth`, `accessTokenCookie`, `refreshTokenCookie`.
- SSE resume dùng parameter `LastEventId` (`Last-Event-ID`).
- Chưa có endpoint nghiệp vụ trong bootstrap; mỗi contract được thêm cùng workflow đã duyệt.
- Quy ước đầy đủ: `littlecafe-docs/api/conventions.md`.
