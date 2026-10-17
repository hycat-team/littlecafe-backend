# LittleCafe Backend Constitution

## I Nguồn chuẩn và workflow

Requirement, business rule, workflow và ADR dùng repository littlecafe-docs. Machine-readable contract trong contracts/ của backend là nguồn tích hợp chuẩn. Source và test là bằng chứng implementation, không được dùng để tự thay đổi intent đã duyệt. Feature chỉ implement sau khi quyết định phụ thuộc đã Confirmed. Khi lựa chọn còn thiếu làm đổi public contract, data model, kiến trúc hoặc hành vi user-visible, agent MUST dừng và hỏi thay vì tự chốt.

## II Ranh giới service

Bảy service là Identity, Order, Activity, Wallet, AI, Notification và Reporting; Content thuộc Activity. Mỗi service sở hữu database và migration riêng. Cấm import implementation hoặc truy cập bảng của service khác. Read xuyên service dùng gRPC; mutation dùng RabbitMQ event.

## III Kiến trúc và contract

Mỗi service dùng Clean Architecture và Wire. OpenAPI contract-first với Gin và oapi-codegen strict server interface. Event dùng AsyncAPI/JSON Schema; protobuf chỉ dành cho read-only cross-service contract. Generated code được commit và CI kiểm tra drift. Không tạo entity, endpoint, RPC hoặc event mẫu.

## IV Dữ liệu và consistency

Identifier dùng UUIDv7. Money dùng int64/BIGINT VND. Timestamp dùng TIMESTAMPTZ; business calendar theo Asia/Ho_Chi_Minh. Mutation quan trọng dùng Idempotency-Key. Transactional outbox, inbox/dedup, retry và DLQ là bắt buộc cho event. Order hoặc Activity sở hữu saga khởi tạo; Notification và Reporting choreography.

## V Security và privacy

Email/password dùng Argon2id; Google OAuth/OIDC được hỗ trợ. Access token là ES256; refresh token là opaque random Base64URL và chỉ lưu digest, có rotation/revoke/reuse detection. Service-to-service dùng mTLS; user JWT được truyền riêng khi hành động thay người dùng. RBAC kết hợp ownership, Guardian-Child và assignment. Dữ liệu nhạy cảm dùng AES-GCM; bank data chỉ là snapshot của withdrawal. PII không xuất hiện trong log, event hoặc fixture ngoài mức tối thiểu.

## VI API và realtime

REST prefix /api/v1. Lỗi là RFC 9457 Problem Details với code, traceId và field errors. Cursor pagination mặc định; offset chỉ cho admin table. SSE có REST fallback và Redis Pub/Sub relay. Redis không phải nguồn dữ liệu authoritative.

## VII Runtime và chất lượng

Một VPS Docker Compose, GHCR pin theo Git SHA. Production migration job chạy trước rollout với advisory lock. Secret và certificate nằm ngoài Git. Zap JSON, Prometheus và OpenTelemetry là chuẩn observability. Mỗi PR chạy format, vet, race test, contract lint, codegen drift, build và kiểm tra boundary phù hợp phạm vi thay đổi.

## VIII Decision mở

3D asset storage, payout entitlement, retention, performance, RPO/RTO, accessibility, AI provider/model/vector store/retrieval policy và các business policy được đánh dấu Undecided trong docs. Scaffold, sample code và dependency có sẵn không phải là quyết định. Agent không được tự chốt.

## Governance

Thay đổi nguyên tắc cần ADR thay thế, cập nhật traceability và constitution version.

Feature thay đổi HTTP endpoint, event payload hoặc DTO client-facing MUST có `specs/<feature>/frontend-guide.md`. Guide dùng một trong các trạng thái `DRAFT`, `IN_PROGRESS`, `READY_FOR_FRONTEND`, `CHANGED`, `DEPRECATED`; tham chiếu contract canonical thay vì chép lại schema; mô tả before → after; và chỉ được đặt `READY_FOR_FRONTEND` sau khi contract ổn định, implementation cùng test evidence đầy đủ và route/field/enum/status đã được đối chiếu có số lượng. Feature không đổi client surface không tạo guide nhưng MUST ghi kết luận trong `tasks.md`.

Finding thật sự ngoài feature MUST đi vào `specs/<feature>/deferred.md` với vấn đề, lý do ngoài phạm vi, điều kiện gỡ và task đã tick bị ảnh hưởng. In-scope work MUST NOT bị chuyển sang deferred để làm sạch exit condition.

Mặc định agent không commit, push, tag hoặc thay đổi branch/history. Khi user gọi `$speckit-orchestrate`, orchestrator có standing authorization tạo một Conventional Commit cho mỗi phase đã qua gate và commit `fixup!` đã verify, chỉ gồm file thuộc phase. Worker không được commit. Push, tag, branch mutation, rebase, reset, amend, squash, migration dữ liệu sống và production action luôn cần phép riêng.

**Version:** 1.1.0 | **Ratified:** 2026-10-17 | **Last Amended:** 2026-10-17

