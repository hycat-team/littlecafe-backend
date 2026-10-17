# Tài liệu kỹ thuật Backend

Tài liệu nghiệp vụ chuẩn nằm trong repository `littlecafe-docs`. Repository này chỉ lưu kiến trúc triển khai, contract machine-readable, hướng dẫn phát triển và vận hành.

Đọc theo thứ tự: [kiến trúc](system-design/architecture.md), [contract](api/contracts.md), [phát triển local](development/local-setup.md), [vận hành](operations/deployment.md).

Feature thay đổi HTTP endpoint, event payload hoặc DTO client-facing phải bàn giao qua `specs/<feature>/frontend-guide.md`. Guide tham chiếu contract canonical, lưu trạng thái bàn giao và được đồng bộ lại khi close-out; feature không đổi client surface không tạo guide nhưng phải ghi kết luận đó trong `tasks.md`.

