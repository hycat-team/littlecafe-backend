# Tài liệu kỹ thuật Backend

Tài liệu nghiệp vụ chuẩn nằm trong repository `littlecafe-docs`. Repository này
lưu kiến trúc triển khai, contract machine-readable, hướng dẫn phát triển và vận
hành.

## Đọc theo task

| Task | Đọc |
|---|---|
| Spec Kit / feature | [spec-kit-workflow](development/spec-kit-workflow.md), [gates](development/gates.md) |
| Git / PR / release | [git-workflow](development/git-workflow.md) |
| Config / env / secret | [environments](development/environments.md) |
| Deploy / rollback | [deployment](operations/deployment.md), [certificates](operations/certificates.md) |
| Kiến trúc / boundaries | [architecture](system-design/architecture.md) |
| Error / handler / DTO / i18n | [design-pattern](system-design/design-pattern.md) |
| Event / outbox / DLQ | [event-driven-design](system-design/event-driven-design.md) |
| Database / migration | [data-and-migrations](system-design/data-and-migrations.md) |
| Security | [security](system-design/security.md) |
| Runtime / observability | [runtime-observability](system-design/runtime-observability.md) |
| Core flows | [core-flows](system-design/core-flows.md) |
| AI service | [ai-service](system-design/ai-service.md) |

## Development

- [spec-kit-workflow](development/spec-kit-workflow.md) — team model, feature identity, helper, Spec Kit phases.
- [git-workflow](development/git-workflow.md) — branch model, PR, release/tag, hotfix.
- [gates](development/gates.md) — nguồn chuẩn duy nhất cho gate.
- [environments](development/environments.md) — config/env/secret, local setup, kiểm tra môi trường.
- [spec-kit-customizations](development/spec-kit-customizations.md) — deviation và upgrade.

## Operations & system design

- [deployment](operations/deployment.md), [certificates](operations/certificates.md).
- [architecture](system-design/architecture.md), [design-pattern](system-design/design-pattern.md),
  [event-driven-design](system-design/event-driven-design.md), [security](system-design/security.md),
  [data-and-migrations](system-design/data-and-migrations.md),
  [runtime-observability](system-design/runtime-observability.md),
  [core-flows](system-design/core-flows.md), [ai-service](system-design/ai-service.md).

Feature thay đổi client surface phải bàn giao qua
`specs/<owner>-<NNN>-<slug>/frontend-guide.md`; feature không đổi client surface
ghi kết luận trong `tasks.md`.
