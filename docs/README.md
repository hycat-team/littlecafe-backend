# Tài liệu kỹ thuật Backend

Tài liệu nghiệp vụ chuẩn nằm trong repository `littlecafe-docs`. Repository này chỉ lưu kiến trúc triển khai, contract machine-readable, hướng dẫn phát triển và vận hành.

Đọc theo thứ tự: [kiến trúc](system-design/architecture.md), [contract](api/contracts.md), [phát triển local](development/local-setup.md), [môi trường](development/environments.md), [vận hành](operations/deployment.md).

## System design

- [Kiến trúc](system-design/architecture.md) — workspace, topology, layout module, layer, routing, giao tiếp liên service.
- [Design patterns](system-design/design-pattern.md) — error/RFC 9457, handler, repository/UoW, Wire, worker, rate limit, SSE, logging, error code.
- [Event-driven design](system-design/event-driven-design.md) — outbox, inbox, DLQ, consumer semantics, metrics.
- [Bảo mật](system-design/security.md) — auth, ES256, refresh token, mTLS, RBAC, dữ liệu nhạy cảm.
- [Service boundaries](system-design/service-boundaries.md) — ranh giới 7 service, contract map, cấm import chéo.
- [Data & migrations](system-design/data-and-migrations.md) — quyền sở hữu dữ liệu, Goose, quy ước entity/money/time.
- [Runtime & observability](system-design/runtime-observability.md) — môi trường chạy, CI/CD, logging, metrics.
- [Core flows](system-design/core-flows.md) — ánh xạ workflow nghiệp vụ sang luồng service.
- [AI service](system-design/ai-service.md) — ranh giới AI, quyết định đã chốt và mâu thuẫn nguồn.

Feature thay đổi HTTP endpoint, event payload hoặc DTO client-facing phải bàn giao qua `specs/<feature>/frontend-guide.md`. Guide tham chiếu contract canonical, lưu trạng thái bàn giao và được đồng bộ lại khi close-out; feature không đổi client surface không tạo guide nhưng phải ghi kết luận đó trong `tasks.md`.
