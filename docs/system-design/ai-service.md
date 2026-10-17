# AI Service

## 1. Ranh giới

- AI là service độc lập, database riêng `littlecafe_ai` (extension `vector` bật sẵn trong `database/init/00-init.sql`).
- AI **không thực hiện mutation** nghiệp vụ; chỉ trả lời/retrieval và dẫn người dùng tới chức năng chính (theo `littlecafe-docs/workflows/rag-assistant.md`).
- Published content được replicate sang pgvector của AI; private/live data đọc qua gRPC read-only sau authorization.
- Không service nào gọi trực tiếp database AI; contract là OpenAPI/AsyncAPI/protobuf như mọi service khác.

## 2. Quyết định đã chốt

| Nguồn | Nội dung |
|---|---|
| `littlecafe-docs/architecture/decisions/001-architecture-decisions.md` (Confirmed) | RAG hybrid với pgvector; Gemini provider interface; Cloudinary cho media 2D |
| `littlecafe-docs/integrations/rag-assistant.md` | Gemini đứng sau provider interface; **chưa có fallback đa nhà cung cấp**; RAG không mutation |
| `.specify/memory/constitution.md` mục IV/VIII | AI service độc lập; dữ liệu nhạy cảm không rời hệ thống; quyết định mở không được tự chốt |

## 3. Mâu thuẫn nguồn cần chốt (agent không tự quyết)

Constitution mục VIII liệt kê **"AI provider/model/vector store/retrieval policy"** là
Undecided, trong khi ADR-001 (Confirmed) và `integrations/rag-assistant.md` đã chốt
Gemini provider interface + pgvector. Hai nguồn đang mâu thuẫn.

- Xử lý đúng: dừng ở phần phụ thuộc, ghi rõ ID/phạm vi ảnh hưởng và hỏi developer.
- Không được tự chọn provider/model/vector store/retrieval policy hay tạo adapter AI
  trước khi mâu thuẫn này được cập nhật ở nguồn chuẩn.

## 4. Chưa chốt (theo nguồn hiện tại)

- Freshness SLA, evaluation threshold, conflict behavior của RAG (`integrations/rag-assistant.md`).
- Fallback khi provider lỗi/rate limit.
- Storage asset 3D (ADR-001: Undecided; WF-007: storage/device Undecided).
- Payout entitlement (ảnh hưởng gián tiếp qua Wallet, không thuộc AI).

## 5. Trạng thái code

- `services/ai` có layout Clean Architecture đầy đủ nhưng module rỗng: không entity,
  use case, endpoint, RPC hay event.
- `shared` chưa có adapter hay config AI (không có `ai.*` trong `configs/base.yaml`).
- Đã có sẵn nền tảng dùng lại: `entities.Vector` (map cột `vector`), `JSONField`,
  outbox/consumer framework, gRPC bridge.

## 6. Điều kiện để bắt đầu implement

1. Chốt mâu thuẫn §3 và cập nhật constitution/ADR cho khớp.
2. Có requirement/workflow RAG được duyệt kèm acceptance criteria.
3. Có OpenAPI (HTTP) + AsyncAPI (event) + protobuf (read-only cross-service) contract tương ứng.
4. Migration bảng vector/ingestion thuộc AI service, không đặt trong service khác.
