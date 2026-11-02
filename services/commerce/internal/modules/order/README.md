# Module order

Module này tuân theo Clean Architecture với bốn lớp `domain`, `application`, `infrastructure` và `presentation`.

Chưa có entity, use case, endpoint, RPC hoặc event nghiệp vụ. Chỉ bổ sung chúng sau khi workflow, quyền sở hữu dữ liệu và contract tương ứng được duyệt. Module này chỉ đọc/ghi persistence của chính nó; không truy cập persistence của module khác. Service sở hữu được ghi trong `docs/system-design/architecture.md`.

