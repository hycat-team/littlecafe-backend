# Tài liệu kỹ thuật Backend

Nghiệp vụ, policy và quyết định kiến trúc nằm trong `littlecafe-docs`. Repository
này giữ contract machine-readable, hướng dẫn implementation và vận hành.

## Đọc theo task

| Task | Đọc |
|---|---|
| Spec Kit / feature | [spec-kit-workflow](development/spec-kit-workflow.md), [gates](development/gates.md) |
| Git / PR / release | [git-workflow](development/git-workflow.md) |
| Config / env / secret | [environments](development/environments.md), [sops-keys](development/sops-keys.md) |
| Deploy / rollback | [deployment](operations/deployment.md) |
| Certificate / mTLS | [certificates](operations/certificates.md) |
| Kiến trúc / boundaries | [architecture](system-design/architecture.md) |
| Error / handler / DTO / i18n | [design-pattern](system-design/design-pattern.md) |
| Spec Kit skill overrides | [speckit-overrides](development/speckit-overrides/README.md) |
| Backend integration handoff | [backend-integration-guide-template](../.specify/templates/backend-integration-guide-template.md) |
| Event / outbox / DLQ | [event-driven-design](system-design/event-driven-design.md) |
| Database / migration | [data-and-migrations](system-design/data-and-migrations.md) |
| Security | [security](system-design/security.md) |
| Runtime / observability | [runtime-observability](system-design/runtime-observability.md) |
| Core flows | [core-flows](system-design/core-flows.md) |
| AI service | [ai-service](system-design/ai-service.md) |

## Nguồn chuẩn theo chủ đề

| Chủ đề | Canonical | Ghi chú |
|---|---|---|
| Policy môi trường, gate, trách nhiệm | `littlecafe-docs/deployment/environments.md` | Mức dự án |
| Quyết định kiến trúc | `littlecafe-docs/architecture/decisions/001-architecture-decisions.md` | ADR-001 |
| Config/env, local setup, pitfall | [environments](development/environments.md) | Giá trị nào ở đâu |
| Cập nhật env, SOPS/age key | [sops-keys](development/sops-keys.md) | Workflow + key lifecycle |
| CD, deploy, rollback | [deployment](operations/deployment.md) | VPS/Compose |
| Certificate, JWT, mTLS | [certificates](operations/certificates.md) | Private CA |
| Gate kỹ thuật | [gates](development/gates.md) | Nguồn chuẩn duy nhất |
| Git/PR/release | [git-workflow](development/git-workflow.md) | Branch model |
| Spec Kit | [spec-kit-workflow](development/spec-kit-workflow.md) | Team model |
| Runbook VPS | `local-docs/deploy/` | Ngoài repository |

## Quy tắc tài liệu

1. Một chủ đề chỉ có một file canonical; file khác tối đa 1–2 câu rồi link.
2. `littlecafe-docs` là policy/why; backend là implementation/how; `local-docs` là
   runbook. Policy không chứa command.
3. Không lặp bảng, command hoặc contract giữa các file; sửa canonical trước.
4. Link trong repo dùng đường dẫn tương đối; link sang repo khác dùng path trong
   backtick (``littlecafe-backend/docs/...``).
5. Độ dài mục tiêu: README ≤ 60 dòng; policy ≤ 100; implementation/ops ≤ 250;
   ADR ≤ 180. Ưu tiên bảng và checklist thay vì đoạn văn.

Feature thay đổi client surface phải bàn giao qua
`specs/<owner>-<NNN>-<slug>/frontend-guide.md`; feature không đổi client surface
ghi kết luận trong `tasks.md`.
