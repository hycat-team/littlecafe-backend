# AI Service

## 1. Ranh giới

- AI là service độc lập, database riêng `littlecafe_ai`; extension `vector` trong `database/init/00-init.sql` chỉ là scaffolding.
- AI **không thực hiện mutation** nghiệp vụ; chỉ retrieval/trả lời và dẫn người dùng tới chức năng chính (WF-008).
- Published content/private data được đọc theo contract; vector store/retrieval policy **Undecided** (OQ-012), chưa chốt replication hay live read.
- Không service nào gọi trực tiếp database AI; contract là OpenAPI/AsyncAPI/protobuf như mọi service khác.

## 2. Quyết định đã chốt

| Nguồn | Nội dung |
| --- | --- |
| `littlecafe-docs/architecture/decisions/001-architecture-decisions.md` (DEC-011) | AI provider/model/vector store/retrieval policy **Undecided**; không cam kết provider cụ thể |
| `littlecafe-docs/integrations/rag-assistant.md` | RAG read-only; source/version; missing/conflict trả "không chắc"; không mutation |
| `littlecafe-docs/workflows/rag-assistant.md` | WF-008 actor/flow/AC |
| `.specify/memory/constitution.md` mục VIII | AI provider/model/vector store/retrieval policy là open decision |
| DEC-020 | Manual evaluation; published content refresh ≤ 15 phút; live private data real-time qua gRPC; conflict behavior |
| DEC-021 | RAG access boundary + privacy/retention |

## 3. Chưa chốt

- Provider/model/vector store/retrieval policy (OQ-012).
- Fallback khi provider lỗi/rate limit.
- Ingestion/replication strategy và vector schema chi tiết.
- Evaluation threshold tự động (hiện manual theo DEC-020).
- Payout entitlement không thuộc AI; theo DEC-018.

## 4. Trạng thái code

- `services/ai` có layout Clean Architecture đầy đủ nhưng module rỗng: không entity, use case, endpoint, RPC hay event.
- `shared` chưa có adapter/config AI (không có `ai.*` trong `configs/base.yaml`).
- Đã có sẵn nền tảng dùng lại: `entities.Vector` (map cột `vector`), `JSONField`, outbox/consumer framework, gRPC bridge. Đây là scaffolding, không phải cam kết provider/retrieval.

## 5. Điều kiện để bắt đầu implement

1. Chốt OQ-012 (provider/model/vector store/retrieval policy) và cập nhật DEC/ADR.
2. Contract OpenAPI/AsyncAPI/protobuf cho RAG.
3. Requirement/workflow RAG đã duyệt (hiện có WF-008, BR-050..BR-059).
4. Migration bảng vector/ingestion thuộc AI service, không đặt trong service khác.
5. Evaluation scenario/source set theo DEC-020.

## 6. Liên kết

- `littlecafe-docs/workflows/rag-assistant.md`
- `littlecafe-docs/integrations/rag-assistant.md`
- `littlecafe-docs/requirements/NFR.md`
