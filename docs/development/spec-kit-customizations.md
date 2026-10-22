# Spec Kit Customizations

Các chỗ lệch có chủ đích so với bản phát hành Spec Kit. Khi upgrade, đối chiếu
bảng này trước khi dùng `--force`.

## 1. Managed files đã sửa

| File | Sửa gì |
|---|---|
| `.agents/skills/speckit-specify/SKILL.md` | dùng `team-feature.ps1 new`; team branch rule |
| `.agents/skills/speckit-analyze/SKILL.md` | block `Feature Binding (team)` |
| `.agents/skills/speckit-checklist/SKILL.md` | block `Feature Binding (team)` |
| `.agents/skills/speckit-clarify/SKILL.md` | block `Feature Binding (team)` |
| `.agents/skills/speckit-converge/SKILL.md` | block `Feature Binding (team)` |
| `.agents/skills/speckit-implement/SKILL.md` | block `Feature Binding (team)` |
| `.agents/skills/speckit-plan/SKILL.md` | block `Feature Binding (team)` |
| `.agents/skills/speckit-tasks/SKILL.md` | block `Feature Binding (team)` |
| `.agents/skills/speckit-taskstoissues/SKILL.md` | block `Feature Binding (team)` |

## 2. File repo tự thêm

| File | Vai trò |
|---|---|
| `.agents/skills/speckit-orchestrate/SKILL.md` | điều phối analyze → implement → converge, git/gate policy |
| `.agents/skills/speckit-worker/SKILL.md` | worker một skill/phase, không git, không sửa hạ tầng |
| `scripts/speckit/team-feature.ps1` | định danh feature: `owner/resolve/new/bind/list/check` |
| `scripts/speckit/tests/team-feature.tests.ps1` | smoke test helper trên temp git repo/clone |
| `.specify/templates/overrides/spec-template.md` | `Feature Branch` → `feat/[slug]` |
| `.specify/templates/overrides/plan-template.md` | thêm "Shared Files & Conflict Hotspots" cho Go |
| `.specify/templates/overrides/tasks-template.md` | tasks template team-mode, tiếng Anh |
| `.specify/templates/deferred-template.md` | repo-owned, không nằm trong manifest |
| `.specify/templates/frontend-guide-template.md` | repo-owned; dùng path `specs/<owner>-<NNN>-<slug>/` |
| `.specify/owner.json` | handle máy-local (gitignored) |
| `specs/<owner>-<NNN>-<slug>/meta.json` | nguồn sự thật của feature |

`checklist-template.md`, `constitution-template.md` và core `tasks-template.md`
không có deviation; core `tasks-template.md` khớp hash manifest 1.0.5. Bản custom
nằm ở `overrides/tasks-template.md`.

## 3. Đánh số và template

- `init-options.json` vẫn ghi `"feature_numbering": "sequential"`, nhưng
  `speckit-specify` không dùng: số do `team-feature.ps1` cấp theo owner.
- Không dùng `create-new-feature.ps1` (đánh số toàn repository).
- Không có `specs/registry.md`; `meta.json` là nguồn sự thật, xem bảng bằng
  `team-feature.ps1 list [-Json]`.
- Override template dùng cơ chế `Resolve-TemplateContent` (ưu tiên
  `.specify/templates/overrides/`).

## 4. Upgrade

1. Đọc diff bản mới cho các file ở mục 1–2.
2. Chép lại phần tuỳ biến nếu bản mới ghi đè.
3. Chạy `specify integration upgrade`; chỉ dùng `--force` sau khi đọc diff.
4. Chạy `pwsh scripts/speckit/team-feature.ps1 check -Strict`,
   `pwsh scripts/speckit/tests/team-feature.tests.ps1`, `make check`.
5. Cập nhật lại bảng hash/nội dung nếu cần.

Không chạy upgrade trên nhánh feature; dùng `chore/*`.
