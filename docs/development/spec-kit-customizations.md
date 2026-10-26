# Spec Kit Customizations

Các chỗ lệch có chủ đích so với bản phát hành Spec Kit. Khi upgrade, đối chiếu
tài liệu này trước khi dùng `--force`.

## 1. Managed files

Tất cả file managed trong `.specify/integrations/*.manifest.json` phải khớp
upstream:

- các managed skill trong `.agents/skills/**` (Codex integration)
- `.specify/templates/*-template.md`
- `.specify/scripts/**`

`specify integration status` phải báo:

```text
Modified managed files: 0
Missing managed files: 0
```

Không sửa tay managed skill/template. Nếu cần đổi hành vi, dùng một trong hai
cơ chế dưới đây.

## 2. Cơ chế override

Spec Kit chính thức hỗ trợ:

1. **Template override**: `.specify/templates/overrides/`.
2. **Command/skill override**: preset (`specify preset add/update`).

LittleCafe không dùng preset và giữ `.agents/skills/**` pristine. Team-specific
behavior nằm trong một **policy overlay** của repo:

```text
docs/development/speckit-overrides/
```

Đây không phải cơ chế chính thức của Spec Kit; enforcement dựa vào constitution
và `AGENTS.md`.

### 2.1 Template overrides

| File | Nội dung |
|---|---|
| `.specify/templates/overrides/plan-template.md` | thêm "Shared Files & Conflict Hotspots" cho Go |
| `.specify/templates/overrides/spec-template.md` | `Feature Branch` → `feat/[slug]` |
| `.specify/templates/overrides/tasks-template.md` | tasks template team-mode, tiếng Anh, frontend/backend handoff decisions |

Managed template vẫn pristine; resolver ưu tiên `.specify/templates/overrides/`.

### 2.2 Skill deltas

| File | Dùng cho |
|---|---|
| `docs/development/speckit-overrides/README.md` | precedence, load order, delta-only rule |
| `feature-binding.md` | mọi managed `speckit-*` |
| `specify.md` | `speckit-specify` |
| `implement.md` | `speckit-implement` |
| `converge.md` | `speckit-converge` |
| `tasks.md` | `speckit-tasks` |
| `handoff.md` | implement/converge/tasks |

Quy tắc:

- Chỉ chứa `Add before step X`, `Replace step X`, `Add after step X`.
- Không copy lại toàn bộ skill.
- Managed skill vẫn là nguồn chính cho các bước không bị override.
- Agent đọc `README.md` + `feature-binding.md` + delta tương ứng trước khi chạy
  managed skill.

## 3. File repo tự thêm

| File | Vai trò |
|---|---|
| `.agents/skills/speckit-orchestrate/SKILL.md` | điều phối analyze → implement → converge, git/gate policy |
| `.agents/skills/speckit-worker/SKILL.md` | worker một skill/phase, không git |
| `docs/development/speckit-overrides/**` | policy overlay delta cho managed skills |
| `.specify/templates/backend-integration-guide-template.md` | handoff gRPC/event/service-to-service |
| `.specify/templates/frontend-guide-template.md` | handoff HTTP/event/DTO client-facing |
| `.specify/templates/deferred-template.md` | ghi việc ngoài scope |
| `scripts/speckit/team-feature.ps1` | định danh feature: `owner/resolve/new/bind/list/check` |
| `scripts/speckit/tests/team-feature.tests.ps1` | smoke test helper |
| `.specify/owner.json` | handle máy-local, gitignored |
| `specs/<owner>-<NNN>-<slug>/meta.json` | nguồn sự thật của feature |

## 4. Đánh số và template

- `init-options.json` vẫn ghi `"feature_numbering": "sequential"`, nhưng
  `speckit-specify` không dùng: số do `team-feature.ps1` cấp theo owner.
- Không dùng `create-new-feature.ps1`.
- Không có `specs/registry.md`; `meta.json` là nguồn sự thật.
- Override template dùng `Resolve-TemplateContent`, ưu tiên
  `.specify/templates/overrides/`.
- Template mới repo-owned đặt trong `.specify/templates/`.

## 5. Upgrade

1. `specify integration status` phải 0 modified.
2. Đọc diff bản Spec Kit mới nếu có.
3. Chạy `specify integration upgrade`.
4. Nếu managed file bị đổi:
   - không chép tay vào `.agents/skills/**`;
   - cập nhật delta tương ứng trong `docs/development/speckit-overrides/`;
   - ghi lại thay đổi trong tài liệu này.
5. Chạy `pwsh scripts/speckit/team-feature.ps1 check -Strict`,
   `make team-check`, `make check`.

Không chạy upgrade trên nhánh feature; dùng nhánh `chore/*`.
