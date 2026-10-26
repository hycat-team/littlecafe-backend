# Git Workflow & Releases

Nguồn chuẩn cho branch model, PR và release của LittleCafe Backend. Quy tắc agent
với Git nằm trong [`spec-kit-workflow.md`](spec-kit-workflow.md) §9.

## 0. Branch preflight trước khi sửa

Trước mọi thao tác làm thay đổi repo - write/edit file, codegen, format,
migration, config/env/deploy, cài dependency, `git add`/commit - agent MUST kiểm
tra branch hiện tại:

```bash
git branch --show-current
```

- **Protected**: `main`, `master`, `develop`, `trunk`, detached HEAD.
- Trên protected branch:
  - Không sửa file.
  - Hỏi developer bằng `question` tool.
  - Chỉ tạo/switch branch sau khi được duyệt.
- **Read-only** được phép trên mọi branch: đọc, search, review, trả lời câu hỏi.
- Không tự tạo/switch/stash branch.
- Nếu branch hiện tại không phải branch của task (kể cả đang ở feature branch
  khác), hỏi trước khi sửa.
- Nếu working tree dirty, hỏi:
  1. Tạo branch từ HEAD hiện tại để giữ thay đổi
  2. Stash → switch `develop` → tạo branch → pop
  3. Dừng
- Nếu developer chủ động yêu cầu làm trên `main`/`develop`, agent phải cảnh báo
  và yêu cầu xác nhận rõ ràng; mặc định vẫn đề xuất tách branch.

Base branch:

- `feat/<slug>`, `fix/<slug>`, `chore/<slug>`: từ `develop`.
- `hotfix/<slug>`: từ `main`.

## 1. Vai trò các nhánh

| Branch | Vai trò | Quy tắc |
|---|---|---|
| `main` | Production/release | Protected; không commit trực tiếp; chỉ nhận `develop`/`hotfix/*`; tag `v*` |
| `develop` | Integration/staging | Protected; không commit trực tiếp; push trigger CD staging |
| `feat/<slug>` | Feature | Tạo từ `develop`; PR về `develop`; 1 owner |
| `fix/<slug>` | Bug fix trên develop | Tạo từ `develop`; PR về `develop` |
| `chore/*` | Hạ tầng/skills/docs | Tạo từ `develop`; PR về `develop` |
| `hotfix/<slug>` | Sửa gấp production | Tạo từ `main`; PR về `main`; tag; back-merge `develop` |

## 2. Flow

```text
feat/<slug>|fix/<slug>|chore/* --PR--> develop --push--> CI + CD staging --> UAT
develop --PR/FF--> main --tag v*--> CD production
hotfix/<slug> --PR--> main --tag--> CD production
main --back-merge--> develop
```

- CI chạy trên PR/push vào `main`/`develop`.
- Staging deploy khi push `develop` hoặc dispatch; docs-only bỏ qua.
- Production chỉ deploy từ tag `v*` hoặc dispatch; **không** deploy khi push
  `main`.
- Không commit trực tiếp lên `main`/`develop`.

## 3. Tag và release

- Chỉ tag trên `main`, format `vX.Y.Z`.
- Không di chuyển hoặc xoá tag đã publish.
- Trước khi tag, kiểm tra:
  - `git fetch origin`
  - `git merge-base --is-ancestor origin/develop origin/main`
  - tag commit thuộc `main`.
- Nếu `main` có commit không có trên `develop`, back-merge `main` → `develop`
  trước khi release tiếp.
- Sau release/hotfix, back-merge `main` → `develop` ngay.

### Checklist release

```bash
git fetch origin
git switch develop
git pull --ff-only
make check

git switch main
git merge --ff-only develop
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin main vX.Y.Z

git switch develop
git merge --ff-only main
git push origin develop
```

### Checklist hotfix

```bash
git switch -c hotfix/<slug> main
# fix + test
# PR hotfix/<slug> -> main
# tag vX.Y.Z trên main
# back-merge main -> develop
```

## 4. Drift checks

```bash
git rev-list --left-right --count origin/main...origin/develop
git merge-base --is-ancestor origin/develop origin/main
git branch --contains <tag>
```

- Output `main-only  develop-only`. Trước release, `main-only` phải là `0`;
  `develop-only` có thể `> 0`. `main-only > 0` nghĩa là `main` có commit không có
  trên `develop` — phải back-merge hoặc revert trước khi release.
- `merge-base --is-ancestor origin/develop origin/main` chỉ đúng sau khi `main`
  đã catch up `develop`; trước release có thể sai.
- Tag phải nằm trên `main`.
- Nếu lệch: dừng release, xác định commit main-only, rồi back-merge hoặc revert
  có kiểm soát.

## 5. PR

- `feat/*`, `chore/*` → `develop`.
- `develop`, `hotfix/*` → `main`.
- PR cần 1 reviewer; mô tả nêu feature id (`owner/NNN/slug`), gate result,
  shared files, migration/frontend-guide nếu có.
- Không push hộ người khác; không force-push branch đã có review.
