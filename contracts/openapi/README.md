# OpenAPI

- Server production chuẩn: `https://api.littlecafe.kids/api/v1`.
- Dùng `oapi-codegen` strict server interface với Gin.
- Chạy `make openapi-gen`; file `generated.gen.go` phải được commit và CI kiểm tra drift.
- Mọi mutation quan trọng khai báo header `Idempotency-Key`.
- Lỗi dùng RFC 9457 Problem Details.
- Chưa có endpoint nghiệp vụ trong bootstrap; mỗi contract được thêm cùng workflow đã duyệt.

