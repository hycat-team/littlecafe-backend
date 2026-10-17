# Hướng dẫn agent Backend

## Nguồn chuẩn

- Nghiệp vụ, requirements, workflow và ADR dùng repository `littlecafe-docs`.
- Machine-readable API/event/gRPC contract dùng `contracts/` trong repository này.
- Source và test là bằng chứng implementation; khi mâu thuẫn với contract hoặc tài liệu đã duyệt, dừng và báo rõ phạm vi ảnh hưởng.

## Quy trình task

1. Đọc tài liệu liên quan trong `littlecafe-docs`, sau đó đọc `docs/README.md` và contract liên quan.
2. Không triển khai hành vi phụ thuộc API nếu workflow/contract chưa được duyệt.
3. Giữ ownership: service không import implementation hoặc truy cập database của service khác.
4. Cross-service read dùng protobuf/gRPC; mutation dùng event contract và outbox.
5. Viết test cùng thay đổi; chạy codegen rồi kiểm tra drift.
6. Nếu feature đổi HTTP endpoint, event payload hoặc DTO client-facing, cập nhật `specs/<feature>/frontend-guide.md` và chỉ đặt `READY_FOR_FRONTEND` khi contract, implementation, test evidence và sync check đã hoàn tất.

## Quy ước bắt buộc

- Clean Architecture và Wire; module Content thuộc Activity service.
- OpenAPI contract-first, Gin + `oapi-codegen` strict server interface.
- UUIDv7; tiền là `int64` VND; thời gian PostgreSQL là `TIMESTAMPTZ` và API RFC3339 offset +07:00.
- RFC 9457 Problem Details; cursor pagination mặc định.
- ES256, opaque refresh token digest, Argon2id, mTLS và AES-GCM theo tài liệu security.
- Không tạo entity, route, RPC, event hoặc fixture giả để minh họa.
- Không commit, push, tag, branch, reset hoặc release khi chưa được yêu cầu rõ. Ngoại lệ duy nhất: khi user gọi `$speckit-orchestrate`, skill có standing authorization tạo đúng một commit cho mỗi phase đã qua gate và các commit `fixup!` đã verify. Worker không có quyền này; push và mọi Git mutation khác vẫn phải xin phép.

## Gate

- Thay đổi nhỏ: `make fmt-check`, test package liên quan.
- Thay đổi contract/provider/wiring: chạy `make generate`, lint và test liên quan, rồi kiểm tra Git diff.
- Trước bàn giao lớn: `make check`, contract lint và Compose config/smoke phù hợp.

Tài liệu và báo cáo viết bằng tiếng Việt; identifier kỹ thuật giữ nguyên tiếng Anh. Không ghi absolute local path hoặc secret vào repository.

