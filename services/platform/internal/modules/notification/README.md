# Module notification

Module này tuân theo Clean Architecture với bốn lớp `domain`, `application`, `infrastructure` và `presentation`.

Chưa có entity, use case, endpoint, RPC hoặc event nghiệp vụ. Chỉ bổ sung chúng sau khi workflow, quyền sở hữu dữ liệu và contract tương ứng được duyệt. Module này chỉ đọc/ghi persistence của chính nó; không truy cập persistence của module khác. Service sở hữu được ghi trong `docs/system-design/architecture.md`.

## Localization

- Nội dung gửi ra client (email/SMS/nội dung notification) dùng message key và bundle riêng của service dưới `locales/vi.json`, `locales/en.json`, đăng ký lúc boot qua `i18n.Default().Register(...)`.
- Consumer đọc `EventEnvelope.locale` rồi gọi `i18n.WithLocale(ctx, locale)` trước khi render; thiếu locale thì dùng `vi`.
- Key mới phải có đủ bản `vi`/`en`; test kiểm tra parity trước khi merge.

