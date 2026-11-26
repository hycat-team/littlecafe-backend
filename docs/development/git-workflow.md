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

- **Protected**: `main`, `master`, `trunk`, detached HEAD.
- `develop` là integration branch và vẫn là protected branch; developer cấp
  **standing approval** để commit trực tiếp **chỉ với small chore**:
  - `docs/**`, các file `*.md` ở root (README, AGENTS.md)
  - `.github/**`
  - `Makefile`, `.editorconfig`, `.gitattributes`, `.gitignore`, `.dockerignore`
  - `scripts/**` (trừ `scripts/speckit/**`)
  - Không đụng `services/**`, `shared/**`, `contracts/**`, `database/**`,
    `configs/**`, `deployments/**`, `.env*`, `.sops.yaml`, `specs/**`,
    `.specify/**`, `.agents/**`.
  - Một mục đích; không đổi runtime/contract/DB/secret/deploy logic; CI xanh;
    không tag, không đụng `main`.
  - Nếu chạm cả path được phép và không được phép → tạo `chore/<slug>`.
- Trên `main`/`master`/`trunk`/detached HEAD, hoặc thay đổi không phải small
  chore trên `develop`:
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
- Nếu developer chủ động yêu cầu làm trên `main`, agent phải cảnh báo và yêu
  cầu xác nhận rõ ràng; mặc định vẫn đề xuất tách branch. Trên `develop`, small
  chore được commit trực tiếp; việc khác vẫn tách branch.

Base branch:

- `feat/<slug>`, `fix/<slug>`, `chore/<slug>`: từ `develop`.
- `hotfix/<slug>`: từ `main`.

## 1. Vai trò các nhánh

| Branch | Vai trò | Quy tắc |
|---|---|---|
| `main` | Production/release | Protected; không commit trực tiếp; chỉ nhận `develop`/`hotfix/*`; tag `v*` |
| `develop` | Integration/staging | Small chore được commit trực tiếp (§0); việc khác tạo nhánh từ `develop`; push trigger CD staging |
| `feat/<slug>` | Feature | Tạo từ `develop`; PR về `develop`; 1 owner |
| `fix/<slug>` | Bug fix trên develop | Tạo từ `develop`; PR về `develop` |
| `chore/*` | Hạ tầng/skills/docs | Tạo từ `develop`; PR về `develop` (small chore có thể commit trực tiếp) |
| `hotfix/<slug>` | Sửa gấp production | Tạo từ `main`; PR về `main`; tag; back-merge `develop` |

## 2. Flow

```text
feat/<slug>|fix/<slug>|chore/* --PR--> develop --push--> CI + CD staging --> UAT
develop --(dev request, FF)--> main --tag v*--> CD production
hotfix/<slug> --PR--> main --tag--> CD production
main --back-merge--> develop
```

- CI chạy trên PR vào `main`/`develop` và push `develop`; `main` không có push
  CI (promotion bằng fast-forward từ `develop`).
- Staging deploy khi push `develop` hoặc dispatch; docs/CI-only bỏ qua.
- Production chỉ deploy từ tag `v*` hoặc dispatch; **không** deploy khi push
  `main`.
- Không commit trực tiếp lên `main`; `develop` chỉ nhận small chore trực tiếp
  (theo §0).

### Dependabot

Security updates của Dependabot nhắm default branch (`main`); cấu hình
`.github/dependabot.yml` gộp các bản vá Go thành một PR và tắt version updates.
Dependabot không cập nhật `go.work.sum`.

Khi PR Dependabot đỏ ở `Static checks`:

```bash
gh pr checkout <PR>
make tidy       # lan version sang module dùng chung (contracts/shared)
make generate   # cập nhật go.work.sum nếu lệch
git add -A && git commit -m "chore(deps): sync workspace versions"
git push
```

- `Check module tidy state` fail → chạy `make tidy`.
- `Verify workspace sums are committed` fail → chạy `make generate`.

### Promotion `develop` -> `main`

Chỉ thực hiện khi developer yêu cầu rõ ràng và CI trên `develop` đã xanh.

```bash
git fetch origin
git merge-base --is-ancestor origin/main origin/develop
git push origin develop:main
```

- Fast-forward, không tạo merge commit.
- Không tạo tag ở bước này; tag `v*` trên `main` mới trigger CD production.
- Hotfix vào `main` vẫn qua PR để có CI; sau đó back-merge `main` vào `develop`.

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
