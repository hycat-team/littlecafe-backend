# Spec Kit Workflow (Team)

Nguồn chuẩn cho cách developer, agent và Spec Kit phối hợp trong LittleCafe
Backend. Gate nằm trong [`gates.md`](gates.md); Git/release nằm trong
[`git-workflow.md`](git-workflow.md).

## 1. Khi nào dùng

- **Toàn bộ feature**: gọi `speckit-orchestrate`. Skill chạy analyze → implement
  → converge, sửa artifact trước khi sửa code, gate và commit một lần mỗi phase.
- **Một bước**: gọi trực tiếp `speckit-specify`, `speckit-clarify`,
  `speckit-plan`, `speckit-tasks`, `speckit-analyze`, `speckit-implement`,
  `speckit-converge`, hoặc các skill bổ trợ bên dưới.
- **Task nhỏ** (hotfix, docs, config): không cần Spec Kit nhưng vẫn theo
  invariant trong `AGENTS.md` và gate tương ứng.
- Khi orchestrator chạy, worker chỉ nhận đúng một skill/phase; không tự chọn
  feature, không Git, không hỏi developer trực tiếp, không spawn subagent.

## 2. Mô hình nhóm

1. **1 feature = 1 owner = 1 branch.** Không hai người cùng sửa `tasks.md` của
   một feature. Nếu cần làm chung, tách thành hai feature.
2. **1 service = 1 feature đang mở** nếu feature tạo migration hoặc đổi
   `contracts/`.
3. `.agents/skills/**`, `.specify/**`, `scripts/speckit/**` do lead sở hữu;
   feature branch không sửa các path này.

## 3. Định danh feature

```text
specs/<owner>-<NNN>-<slug>/     ← thư mục feature
feat/<slug>                     ← nhánh duy nhất của feature
```

- `<owner>`: handle máy-local, 2–20 ký tự `[a-z0-9-]`, lưu ở
  `.specify/owner.json` (gitignored, không commit).
- `<NNN>`: số 3 chữ số, đếm riêng theo owner. Hai owner có thể cùng `001`.
- `<slug>`: duy nhất toàn repo, trùng phần sau `feat/`.
- Không dùng `feat/<owner>/<slug>`.

`meta.json` là nguồn sự thật:

```json
{
  "owner": "duck",
  "number": 1,
  "slug": "api-localization",
  "branch": "feat/api-localization",
  "kind": "feat",
  "status": "open",
  "services": ["shared"],
  "description": "API message localization",
  "opened_at": "2026-10-22"
}
```

Chỉ sửa tay `status`, `services`, `description`; `owner`, `number`, `slug`,
`branch` do helper sinh. Không có `specs/registry.md`; dùng `list` để xem bảng.

### Trùng slug

Slug global vì branch là `feat/<slug>`. Developer tạo branch xong thì push ngay
để remote ref giữ chỗ. `new` đọc `origin/feat/<slug>`; nếu đã có `meta.json` của owner
khác, helper dừng. Nếu hai người cùng tạo slug khi chưa push, `make team-check`
trước khi merge sẽ phát hiện; người sau đổi slug + branch + tên thư mục.

## 4. Helper

Mọi thao tác định danh đi qua `scripts/speckit/team-feature.ps1` (không dùng
`create-new-feature.ps1` — script managed đánh số toàn repository):

```bash
pwsh scripts/speckit/team-feature.ps1 owner -Handle duck
pwsh scripts/speckit/team-feature.ps1 owner

pwsh scripts/speckit/team-feature.ps1 resolve -Json
pwsh scripts/speckit/team-feature.ps1 new -Slug user-auth -Description "Email/password auth" -Json
pwsh scripts/speckit/team-feature.ps1 bind

pwsh scripts/speckit/team-feature.ps1 list
pwsh scripts/speckit/team-feature.ps1 list -Json
pwsh scripts/speckit/team-feature.ps1 check
pwsh scripts/speckit/team-feature.ps1 check -Strict
```

Các skill gọi `resolve` trước khi ghi file và STOP nếu nhánh hiện tại không khớp
feature đã bind. `-Strict` biến warning (bind cũ, thư mục legacy, `owner.json`
chưa ignore) thành lỗi. Chạy `make team-check` trước khi mở PR; `make check`
cũng gọi `team-check`.

## 5. Luồng làm việc

1. Đồng bộ: `git fetch origin && git switch develop && git pull --ff-only`.
2. Tạo nhánh: `git switch -c feat/<slug>`.
3. Developer giữ chỗ slug: `git push -u origin feat/<slug>` (nhánh rỗng cũng được).
4. Chạy `$speckit-specify`; nếu chưa có owner handle, agent hỏi rồi ghi hộ.
   Helper `new` sinh thư mục + `meta.json` + bind `feature.json`.
5. Điền `services` trong `meta.json`; khai báo shared files trong `plan.md`.
6. Chạy các phase Spec Kit bên dưới.
7. Trước PR: `make team-check`, cập nhật `status`, chạy `make check`.
8. PR 1 reviewer; sau merge, merge `main` vào nhánh trước khi tạo migration mới.

### 5.1 Managed skills và team overrides

- `.agents/skills/**` và `.specify/templates/*-template.md` là file managed của
  Spec Kit; MUST giữ nguyên bản gốc, không sửa tay.
- Team-specific behavior nằm trong
  [`speckit-overrides/`](speckit-overrides/README.md) dưới dạng delta:
  - `feature-binding.md` cho mọi skill;
  - `specify.md`, `implement.md`, `converge.md`, `tasks.md`, `handoff.md` khi
    skill tương ứng cần.
- Trước khi chạy managed `speckit-*`, agent MUST đọc
  `speckit-overrides/README.md` + `feature-binding.md` + delta tương ứng.
- Managed skill vẫn là nguồn chính cho mọi bước không bị override.
- `speckit-orchestrate` và `speckit-worker` là custom skills repo-owned, được
  phép khác upstream.
- Template custom dùng `.specify/templates/overrides/`; template mới repo-owned
  dùng `.specify/templates/`.

Đây là policy overlay của repo, không phải cơ chế command/skill override chính
thức của Spec Kit. Spec Kit chính thức hỗ trợ template overrides và preset-based
command/skill overrides.

## 6. Các phase

| Phase | Skill | Output / điều kiện dừng |
|---|---|---|
| 0 | Helper `new` / `resolve` | `specs/<owner>-<NNN>-<slug>/`, `meta.json`, bind đúng branch |
| 1 | `speckit-specify` | `spec.md`: user story độc lập, yêu cầu, success criteria |
| 2 | `speckit-clarify` | Làm rõ tối đa 5 điểm mơ hồ, cập nhật lại `spec.md` |
| 3 | `speckit-plan` | `plan.md`: Constitution Check, data model, contract/schema, research |
| 4 | `speckit-tasks` | `tasks.md`: task theo user story, `[P]`, đường dẫn file, frontend/backend handoff |
| 5 | `speckit-analyze` | Read-only; CRITICAL/HIGH phải sửa ở artifact trước khi code |
| 6 | `speckit-implement` | Code theo phase, tick `[x]`, gate scoped |
| 7 | `speckit-converge` | Đối chiếu codebase với spec/plan/tasks, append task còn thiếu |

Chi tiết từng bước nằm trong `.agents/skills/speckit-*/SKILL.md`; doc này chỉ
giữ phase, output và điều kiện dừng.

### Phase 1–4

- `spec.md` lấy nguồn từ requirement/workflow/ADR đã duyệt trong
  `littlecafe-docs`; không tự chốt quyết định mà nguồn đang mở.
- `plan.md` cập nhật contract machine-readable trước implementation.
- `tasks.md` phải nói rõ hai handoff:
  - frontend: có `frontend-guide.md` hay kết luận không đổi client surface;
  - backend: có `backend-integration-guide.md` hay kết luận không đổi
    gRPC/event/service-to-service surface.
- Việc ngoài feature ghi `deferred.md`.

### Phase 6–7

- Chạy gate theo [`gates.md`](gates.md); không chạy `make check` giữa các phase.
- Phase chạm `specs/**` hoặc `meta.json` thì chạy thêm `make team-check`.
- Converge sạch: chạy `make check` một lần, sync `frontend-guide.md` và
  `backend-integration-guide.md` nếu surface tương ứng đổi, rồi commit close-out.

### Skill bổ trợ

- `speckit-checklist`: checklist chất lượng yêu cầu trong `checklists/`.
- `speckit-constitution`: cập nhật `.specify/memory/constitution.md`.
- `speckit-taskstoissues`: chuyển task thành GitHub Issues khi cần theo dõi.

## 7. File dùng chung

Chỉ sửa khi `plan.md` khai báo và không feature đang mở nào khác sở hữu:

| Path | Cách xử lý |
|---|---|
| `contracts/openapi/**`, `contracts/asyncapi/**`, `contracts/proto/**` | khai báo plan; merge tuần tự; feature sau rebase |
| `configs/base.yaml`, `configs/<service>.yaml` | feature sau rebase và thêm lại |
| `.env.example`, `.env.*.example` | cập nhật đồng thời config + Compose + docs |
| `deployments/docker-compose*.yml` | khai báo plan; tránh sửa cùng lúc |
| `services/<service>/**/wire_gen.go` | chỉ regen sau merge; không sửa tay |
| `services/<service>/database/migrations/**` | một service một feature; rebase `main` trước |
| `go.work`, `go.work.sum` | gộp PR nhỏ riêng nếu có thể |
| `Makefile`, `.github/workflows/ci.yml` | lead sở hữu; PR `chore/*` |
| `shared/**` | chỉ thêm khi thật sự cross-cutting |
| `docs/system-design/**` | cập nhật cùng feature nếu kiến trúc đổi |
| `.agents/skills/**`, `.specify/**`, `scripts/speckit/**` | lead sở hữu; PR `chore/*`. Managed skills/templates giữ upstream |
| `docs/development/speckit-overrides/**` | lead sở hữu; PR `chore/*`; delta-only cho managed skills |

## 8. Migration

- Một service chỉ có một feature đang mở được tạo migration.
- Merge `origin/main` trước `make migrate-create svc=<service> name=<name>`.
- Không sửa migration đã áp dụng; nếu lệch, xoá migration vừa tạo và sinh lại.
- Chi tiết Goose/schema: [`../system-design/data-and-migrations.md`](../system-design/data-and-migrations.md).

## 9. Git & PR

- Agent không tự commit/push/tag/branch/reset/rebase/amend/merge nếu chưa được
  yêu cầu; chỉ đọc git.
- Trước mọi file mutation, kể cả task ngoài Spec Kit, áp dụng
  [`git-workflow.md`](git-workflow.md) §0: kiểm tra branch, dừng và hỏi nếu đang
  ở protected branch. Không tự tạo/switch/stash branch.
- `$speckit-orchestrate` có standing authorization cho đúng một commit mỗi phase
  đã qua gate và `fixup!` đã verify; worker không có quyền này.
- Chỉ commit trên branch của task (`feat/<slug>`, `fix/<slug>`,
  `chore/<slug>`, `hotfix/<slug>`); không commit `main`/`develop`,
  `.specify/owner.json`,
  `.specify/feature.json`, hoặc `specs/**` của feature khác.
- Managed `.agents/skills/**` và `.specify/templates/*-template.md` MUST remain
  upstream. Không sửa tay; team delta nằm trong
  [`speckit-overrides/`](speckit-overrides/README.md).
- Conventional Commits, message tiếng Anh. Chỉ stage path thuộc feature.
- Branch model và release/tag: [`git-workflow.md`](git-workflow.md).
- Trước push: `make check`; trước PR: `make team-check`.

## 10. Checklist trước PR

- [ ] `make team-check` xanh, không warning mới.
- [ ] `meta.json`: `services` và `status` đúng.
- [ ] `plan.md` liệt kê đủ shared files; không sửa file feature khác.
- [ ] Diff không có thay đổi ngoài phạm vi feature.
- [ ] `make check` xanh.
- [ ] PR có 1 reviewer; mô tả nêu `owner/NNN/slug`, gate, shared files,
      migration/frontend-guide/backend-integration-guide nếu có.
- [ ] Không có managed skill/template bị sửa tay; delta nằm trong
      `docs/development/speckit-overrides/`.

## 11. Nâng cấp Spec Kit

Skill canonical ở `.agents/skills/<name>/SKILL.md` (trừ `speckit-orchestrate`,
`speckit-worker` do repo tự thêm). Danh sách deviation và quy trình upgrade:
[`spec-kit-customizations.md`](spec-kit-customizations.md). Không chạy upgrade
trên nhánh feature; dùng nhánh `chore/*`.
