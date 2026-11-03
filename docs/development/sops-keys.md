# SOPS và age cho biến môi trường

Canonical cho: quy trình cập nhật env, cấp/thu hồi age key, rotate, recovery.
Phân loại giá trị: [`environments.md`](environments.md); deploy/rollback:
[`../operations/deployment.md`](../operations/deployment.md); policy:
`littlecafe-docs/deployment/environments.md`.

## 1. Mô hình

| Môi trường | Nguồn chuẩn | Ai giải mã |
| --- | --- | --- |
| Local | `.env` per-dev từ `.env.example` | Chỉ máy dev (không có secret thật) |
| Staging | `deployments/env/staging.env.enc` | Dev được cấp, lead, VPS staging |
| Production | `deployments/env/production.env.enc` | Dev được cấp, lead, VPS production, recovery key |

- SOPS mã hóa value; tên key và comment `#` vẫn đọc được để review diff.
- File `.enc` có metadata `sops_*` và MAC; sửa tay sẽ hỏng file.
- Plaintext runtime chỉ tồn tại trên VPS (mode 600), do CD giải mã từ Git.

## 2. Ai giữ key nào

| Key | Nơi lưu |
| --- | --- |
| Lead | `~/.config/sops/age/keys.txt` |
| Dev | như trên, trên máy dev |
| VPS production | `/opt/littlecafe-secrets/age/keys.txt` (600) |
| VPS staging | `/opt/littlecafe-staging-secrets/age/keys.txt` (600) |
| Recovery production | Password manager/giấy in offline |

Trên Windows, SOPS chỉ tự đọc `%AppData%\sops\age\keys.txt`; identity tại
`~/.config` phải được trỏ bằng `SOPS_AGE_KEY_FILE` (script onboarding tự set user
env). Linux/macOS tự đọc `~/.config/sops/age/keys.txt`.

Public key `age1...` an toàn để commit/gửi chat. Private key
`AGE-SECRET-KEY-1...` không commit, không gửi, không dán vào issue/PR/log. File
key phải backup (password manager) và bật full-disk encryption.

## 3. Cài công cụ

```powershell
winget install --id FiloSottile.age -e
winget install --id SecretsOPerationS.SOPS -e
sops --version   # 3.13.x
```

Linux/macOS: `apt/brew install age`; tải `sops` từ GitHub releases v3.13.3.
Không muốn cài: `docker run --rm ghcr.io/getsops/sops:v3.13.3-alpine --version`
(chỉ dùng cho decrypt/updatekeys; `sops edit` cần binary).

## 4. Onboarding

### 4.1. Local (không cần key)

Setup đầy đủ: [`bootstrap.md`](bootstrap.md).

```powershell
make env-local        # tạo .env từ .env.example nếu chưa có
make env-local-check
```

### 4.2. Nhận quyền staging/production

1. `pwsh scripts/env/init-dev-key.ps1` (Linux/macOS:
   `sh scripts/env/init-dev-key.sh`) — tạo `~/.config/sops/age/keys.txt`, set
   `SOPS_AGE_KEY_FILE` nếu chưa có, in public key. Đã có identity khác: dùng
   `-New`/`--new` để thêm.
2. Gửi lead: handle, tên máy, `age1...`. Không gửi private key.
3. Lead thêm vào `.sops.yaml` → `make env-updatekeys` → commit/merge.
4. `git pull`, xác nhận:

```powershell
make env-verify
make env-apply ENV=staging   # tạo .env.staging local khi cần chạy tay
```

Không dùng script: `age-keygen -o <path>` rồi copy vào `keys.txt`; xem docs age.

## 5. Quy trình cập nhật env

### 5.1. Local

Sửa trực tiếp `.env`; không commit, không cần SOPS. Khi `.env.example` thêm key,
chạy `make env-local-check`.

### 5.2. Staging/production — luồng chuẩn

1. Phân loại giá trị theo [`environments.md`](environments.md) §2.1. Key mới phải
   thêm vào `.env.<env>.example` **cùng PR**.
2. Sửa ciphertext:

   ```powershell
   $env:EDITOR = "code --wait"      # hoặc notepad
   make env-edit ENV=staging        # hoặc production
   ```

3. Kiểm tra local:

   ```powershell
   git diff -- deployments/env/staging.env.enc
   make env-check     # key set khớp .example, không thừa/thiếu
   make env-verify    # decrypt bằng key của bạn
   ```

4. Commit trên nhánh từ `develop`, mở PR:

   ```powershell
   git switch -c chore/env-<mô-tả> develop
   git add deployments/env .env.staging.example
   git commit -m "chore(env): <mô tả>"
   git push -u origin chore/env-<mô-tả>
   ```

5. Merge `develop` → CD staging tự chạy (apply env + rollout + smoke). Kiểm tra
   health `https://staging-api.littlecafe.kids:8443/api/v1/health` và CORS.
6. Merge `develop` → `main`, tag `v*` (release) hoặc dispatch `CD - VPS` nếu
   khẩn cấp; CD production apply env + rollout.
7. Kiểm tra health `https://api.littlecafe.kids:8444/api/v1/health`, container,
   log. Không `docker compose down -v`.

### 5.3. Thêm / xóa key

- Thêm: `.example` + cả hai `.enc` trong cùng PR. `make env-check` fail nếu thiếu
  hoặc thừa.
- Xóa: xác nhận compose/code không còn dùng (`grep` trong `deployments/`,
  `configs/`, `services/`), xóa khỏi `.example` và hai `.enc` bằng `make env-edit`.
- `sops unset` không hỗ trợ dotenv flat key; xóa trực tiếp trong editor.

### 5.4. Rotate secret

- Provider key: sửa env → deploy service dùng key (`commerce`, `platform`...).
- `REDIS_PASSWORD`, `RABBITMQ_PASSWORD`: sửa env → recreate service tương ứng.
- `*_DB_PASSWORD`: sửa env → deploy; `db-init` chạy `ALTER ROLE` trước rollout rồi
  service recreate với password mới.
- Offboarding/nghi lộ: xóa recipient + `make env-updatekeys` + **rotate
  credential thật** (commit cũ vẫn giải được bằng key cũ).

### 5.5. Hotfix khi CD không chạy

Chỉ khi khẩn cấp, có ghi nhận:

```bash
scp deployments/env/staging.env.enc littlecafe-vps:/opt/littlecafe-staging/deployments/env/
ssh littlecafe-vps
cd /opt/littlecafe-staging
SOPS_AGE_KEY_FILE=/opt/littlecafe-staging-secrets/age/keys.txt \
  sh deployments/scripts/apply-env.sh \
    deployments/env/staging.env.enc .env.staging .env.staging.example
# recreate service bị ảnh hưởng
```

Sau đó vẫn phải đưa thay đổi vào Git để lần deploy sau không ghi đè.

### 5.6. Rollback env

- Một bước: `apply-env.sh` giữ `.env.<env>.prev`; copy `.prev` trở lại rồi recreate.
- Xa hơn: `git revert` commit ciphertext (hoặc checkout commit cũ) rồi deploy.
- Image: `.env.deploy.prev` hoặc dispatch CD với image cũ.

### 5.7. Cheat sheet

```powershell
make env-edit ENV=staging     # sửa ciphertext
make env-check                # parity, không cần key
make env-verify               # decrypt thử
make env-apply ENV=staging    # tạo .env.staging local
make env-updatekeys           # sync recipient sau khi sửa .sops.yaml
make env-sops-test            # test round-trip apply-env
```

## 6. Key lifecycle

### 6.1. Thêm dev

Thêm `age1...` vào rule staging + production trong `.sops.yaml` (kèm comment tên)
→ `make env-updatekeys` → commit. Không cần deploy lại; dev `git pull` và
`make env-verify`.

### 6.2. Dev đổi máy/mất key

Dev sinh identity mới, gửi public key; lead thêm (giữ/xóa key cũ) → `updatekeys`.
Máy mất cắp: coi như lộ key production → làm 6.3.

### 6.3. Offboarding/nghi lộ

1. Xóa public key khỏi `.sops.yaml` → `make env-updatekeys` → merge.
2. Rotate data key:

   ```bash
   SOPS_MODE=binary sops rotate --input-type dotenv --output-type dotenv -i deployments/env/production.env.enc
   SOPS_MODE=binary sops rotate --input-type dotenv --output-type dotenv -i deployments/env/staging.env.enc
   make env-updatekeys
   ```

3. Rotate credential thật: Google OAuth, payOS, Brevo, Cloudinary,
   `RABBITMQ_PASSWORD`, `REDIS_PASSWORD`, `POSTGRES_ADMIN_PASSWORD`,
   `<SERVICE>_DB_PASSWORD`.
4. Ghi ngày rotate vào ADR hoặc `docs/operations/`.

### 6.4. Key VPS

Sinh keypair mới trên máy tin cậy, đẩy private key đúng path (`umask 077`, 600,
owner `deploy`); thêm public key mới, xóa key cũ trong `.sops.yaml`,
`updatekeys`, deploy lại và kiểm tra health.

### 6.5. Recovery

Recovery key chỉ là recipient production, private key cất offline. Diễn tập mỗi
năm trên máy sạch:

```bash
SOPS_AGE_KEY_FILE=/duong/dan/recovery-prod.txt \
  sops decrypt --input-type dotenv --output-type dotenv \
    deployments/env/production.env.enc > /dev/null
```

Mất toàn bộ key (lead + VPS + recovery) = không khôi phục được ciphertext.

## 7. VPS và CD

- CD upload `.enc`, `.example`, `sops-run.sh`, `apply-env.sh`, rồi chạy step
  "Apply … env from SOPS" trước `deploy-release.sh`.
- `apply-env.sh`: decrypt → validate (CR, `sops_`, key parity) → giữ `.prev` →
  ghi atomic mode 600.
- `sops-run.sh` dùng image pinned digest, `--network none --read-only
  --cap-drop ALL`, mount key read-only; không chạy `sops edit` trên VPS.
- Nâng cấp SOPS: đổi digest trong `sops-run.sh`, `docker pull` trên VPS, deploy
  staging trước.

## 8. Troubleshooting

| Lỗi | Xử lý |
| --- | --- |
| `sops: command not found` | Mở terminal mới/cài lại; hoặc `SOPS_MODE=docker` |
| `no matching creation rules found` | Chạy từ repo root; kiểm tra `path_regex` trong `.sops.yaml` |
| `no key could decrypt` | Identity chưa được cấp hoặc sai `SOPS_AGE_KEY_FILE` |
| Windows không tìm thấy key | Set user env `SOPS_AGE_KEY_FILE` trỏ `~/.config/sops/age/keys.txt`, mở terminal mới |
| `invalid dotenv input line` | File nguồn CRLF hoặc dòng thiếu `=`; chuyển LF |
| `invalid character '#'` | Thiếu `--input-type dotenv`; dùng `make env-verify`/`env-edit` |
| `sops edit` không mở editor | Đặt `EDITOR` (`code --wait`, `notepad`) |
