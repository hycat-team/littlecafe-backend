# SOPS và age cho biến môi trường

Tài liệu này là nguồn chuẩn cho việc **nhận quyền, sửa, cập nhật và thu hồi key**
đối với các file env được mã hóa của LittleCafe. Đọc mục 4 nếu bạn vừa clone
source và cần chạy local; đọc mục 7 nếu bạn là người cấp/thu hồi quyền.

- Quy ước config/env: [`environments.md`](environments.md).
- Luồng triển khai: [`../operations/deployment.md`](../operations/deployment.md).
- Chính sách môi trường: `littlecafe-docs/deployment/environments.md`.

## 1. Mô hình

| Môi trường | File trong Git | Ai giải mã được |
| --- | --- | --- |
| Local | Không có file mã hóa; `.env` per-dev từ `.env.example` | Chỉ máy dev đó (không phải secret thật) |
| Staging | `deployments/env/staging.env.enc` | Dev được cấp quyền, lead, VPS staging |
| Production | `deployments/env/production.env.enc` | Dev được cấp quyền, lead, VPS production, recovery key |

- SOPS mã hóa **value** trong file dotenv; **tên key và comment dạng `#`** vẫn
  đọc được, nên review thấy được key nào thay đổi mà không thấy giá trị.
- File `.enc` có metadata `sops_*` và MAC; sửa tay sẽ làm hỏng file.
- Plaintext runtime chỉ tồn tại trên VPS tại `/opt/littlecafe/.env.production`
  và `/opt/littlecafe-staging/.env.staging` (mode 600), do CD giải mã từ Git.
- Local dev không dùng SOPS và không cần age key.

## 2. Ai giữ key nào

| Key | Nơi lưu | Ghi chú |
| --- | --- | --- |
| Lead | `%AppData%\sops\age\keys.txt` (Windows), `~/.config/sops/age/keys.txt` (Linux/macOS) | Là recipient của staging + production |
| Dev | như trên, trên máy của dev | Chỉ gửi public key cho lead |
| VPS production | `/opt/littlecafe-secrets/age/keys.txt` (600) | Chỉ dùng cho production |
| VPS staging | `/opt/littlecafe-staging-secrets/age/keys.txt` (600) | Chỉ dùng cho staging |
| Recovery production | Password manager/giấy in offline | Chỉ là recipient của production |

Public key có dạng `age1...`, an toàn để commit và gửi qua chat nội bộ.
Private key có dạng `AGE-SECRET-KEY-1...`, **không** commit, **không** gửi ai,
**không** dán vào issue/PR/log.

## 3. Cài công cụ

Windows (không cần quyền admin):

```powershell
winget install --id FiloSottile.age -e
winget install --id SecretsOPerationS.SOPS -e
sops --version   # 3.13.x
```

Linux/macOS:

```bash
sudo apt install age        # hoặc: brew install age
# sops: tải binary từ https://github.com/getsops/sops/releases (v3.13.3)
```

Nếu không muốn cài `sops`, có thể dùng Docker:

```bash
docker run --rm ghcr.io/getsops/sops:v3.13.3-alpine --version
```

`SOPS_MODE=docker` chỉ dùng cho `decrypt`/`updatekeys`; **`sops edit` cần binary**
vì cần editor. Dev nên cài binary để sửa env.

## 4. Lần đầu clone source

### 4.1. Chạy local (không cần key)

```powershell
make env-local        # tạo .env từ .env.example nếu chưa có
make env-local-check  # so key với .env.example
```

`make env-local` không ghi đè `.env` đang có. Local dùng password placeholder,
provider key rỗng; không có secret thật nên không cần mã hóa.

### 4.2. Nhận quyền đọc staging/production

1. Sinh key trên máy bạn:

   ```powershell
   pwsh scripts/env/init-dev-key.ps1
   ```

   Script tạo `%AppData%\sops\age\keys.txt` (nếu chưa có), in ra **public key**.
   Nếu đã có identity từ project khác, script in key hiện có; dùng `-New` để thêm
   identity mới. Trên Linux/macOS: `sh scripts/env/init-dev-key.sh`.

2. Gửi lead thông tin sau (chỉ public key):

   ```text
   handle: <tên bạn>
   machine: <tên máy>
   pubkey: age1...
   ```

3. Lead thêm public key vào `.sops.yaml`, chạy `make env-updatekeys`, commit.
   Sau khi merge, `git pull` và xác nhận:

   ```powershell
   make env-verify
   ```

   Hoặc gọi trực tiếp (phải chỉ rõ format, vì đuôi `.enc` không cho SOPS biết
   đây là dotenv):

   ```powershell
   sops decrypt --input-type dotenv --output-type dotenv deployments/env/staging.env.enc > $null
   sops decrypt --input-type dotenv --output-type dotenv deployments/env/production.env.enc > $null
   ```

4. Khi cần file env để chạy tay:

   ```powershell
   make env-apply ENV=staging      # tạo .env.staging (gitignored, mode 600)
   ```

Backup private key vào password manager. Mất key không mất dữ liệu: sinh key
mới, gửi public key, lead cấp lại; nhưng phải làm trước khi cần dùng.

## 5. Sinh key thủ công (không dùng script)

Windows:

```powershell
New-Item -ItemType Directory -Force "$env:APPDATA\sops\age" | Out-Null
$out = age-keygen
$out | Out-File -Append -Encoding ascii "$env:APPDATA\sops\age\keys.txt"
($out | Select-String '^# public key: (age1\S+)').Matches[0].Groups[1].Value
```

Linux/macOS:

```bash
mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/sops/age"
age-keygen >> "${XDG_CONFIG_HOME:-$HOME/.config}/sops/age/keys.txt"
grep '^# public key:' "${XDG_CONFIG_HOME:-$HOME/.config}/sops/age/keys.txt" | tail -n1
```

Dùng nhiều máy: sinh một identity cho mỗi máy, gửi cả hai public key và ghi rõ
tên máy trong `.sops.yaml`.

## 6. Sửa env hằng ngày

```bash
# từ repo root, với identity đã được cấp quyền
export EDITOR="code --wait"          # Windows PowerShell: $env:EDITOR="code --wait"
make env-edit ENV=staging
git diff -- deployments/env/staging.env.enc
git add deployments/env/staging.env.enc
git commit -m "chore(env): <mô tả>"
```

- Chỉ sửa qua `sops edit` hoặc `make env-edit`; không sửa trực tiếp file `.enc`.
- Thêm setting mới: cập nhật `configs/` nếu là hằng số, cập nhật
  `.env.<env>.example`, rồi thêm key vào `.env.<env>.enc` của từng môi trường.
- Không đưa giá trị thật vào `.env*.example`, PR description hoặc screenshot.
- `git diff` chỉ hiện `ENC[...]` cho value; tên key vẫn hiện để review.

## 7. Cập nhật key (onboarding, rotate, offboarding)

Mọi thay đổi recipient đều sửa `.sops.yaml` (chỉ public key) rồi chạy:

```bash
make env-updatekeys      # sops updatekeys cho production + staging
git add .sops.yaml deployments/env/*.env.enc
git commit -m "chore(env): update SOPS recipients"
```

`updatekeys` chỉ re-wrap data key, **không đổi secret** và **không cần deploy lại**
service. Người đã bị xóa không đọc được bản mới, nhưng vẫn đọc được commit cũ —
vì vậy khi thu hồi quyền production phải rotate credential thật (mục 7.3).

### 7.1. Thêm dev

1. Nhận `age1...` + handle + tên máy.
2. Thêm vào danh sách `age` của rule staging và production trong `.sops.yaml`,
   kèm comment tên dev.
3. `make env-updatekeys` → commit → merge.
4. Dev `git pull` và xác nhận decrypt (mục 4.2 bước 3).

### 7.2. Dev đổi máy hoặc mất key

1. Dev sinh identity mới trên máy mới, gửi public key mới.
2. Lead thêm key mới; có thể giữ key cũ nếu máy cũ còn dùng, hoặc xóa nếu mất máy.
3. `make env-updatekeys` → commit.
4. Nếu máy cũ bị mất cắp: coi như key production bị lộ, làm mục 7.3.

### 7.3. Offboarding / nghi lộ key

1. Xóa public key khỏi `.sops.yaml` (staging + production).
2. `make env-updatekeys` → commit → merge.
3. Rotate data key:

   ```bash
   SOPS_MODE=binary sops rotate --input-type dotenv --output-type dotenv -i deployments/env/production.env.enc
   SOPS_MODE=binary sops rotate --input-type dotenv --output-type dotenv -i deployments/env/staging.env.enc
   make env-updatekeys
   ```

4. Rotate credential thật vì commit cũ vẫn giải được bằng key cũ:
   - Provider: Google OAuth client secret, payOS keys, Brevo API key, Cloudinary.
   - Hạ tầng: `RABBITMQ_PASSWORD`, `REDIS_PASSWORD`, `POSTGRES_ADMIN_PASSWORD`,
     `<SERVICE>_DB_PASSWORD` (chạy deploy để `db-init` cập nhật role password).
5. Ghi ngày rotate + lý do vào ADR hoặc `docs/operations/`.

### 7.4. Đổi key của VPS

1. Sinh keypair mới trên máy tin cậy, đẩy private key lên đúng path:
   - production: `/opt/littlecafe-secrets/age/keys.txt`
   - staging: `/opt/littlecafe-staging-secrets/age/keys.txt`
   với `umask 077`, mode 600, owner `deploy`.
2. Thêm public key mới vào `.sops.yaml`, xóa public key cũ, `make env-updatekeys`,
   commit.
3. Deploy lại stack liên quan và kiểm tra health; nếu decrypt lỗi, khôi phục key
   cũ và `.prev` của env file.

### 7.5. Recovery

- Recovery keypair chỉ là recipient của production; private key cất offline.
- Diễn tập mỗi năm một lần trên máy sạch:

  ```bash
  SOPS_AGE_KEY_FILE=/duong/dan/recovery-prod.txt \
    sops decrypt --input-type dotenv --output-type dotenv \
      deployments/env/production.env.enc > /dev/null
  ```

- Nếu mất toàn bộ key (lead + VPS + recovery), không thể khôi phục ciphertext
  trong Git; phải tạo lại env từ backup/credential mới.

## 8. VPS và CD

- CD upload `deployments/env/<env>.env.enc`, `.env.<env>.example`,
  `deployments/scripts/{sops-run.sh,apply-env.sh}`.
- Step "Apply env from SOPS" chạy trước `deploy-release.sh`:

  ```bash
  cd /opt/littlecafe
  SOPS_AGE_KEY_FILE=/opt/littlecafe-secrets/age/keys.txt \
    sh deployments/scripts/apply-env.sh \
      deployments/env/production.env.enc .env.production
  ```

- `sops-run.sh` dùng image pinned digest
  `ghcr.io/getsops/sops:v3.13.3-alpine@sha256:...`, chạy `--network none
  --read-only --cap-drop ALL`, mount key read-only.
- Không chạy `sops edit` trên VPS. Không sửa `.env.production` bằng tay; mọi thay
  đổi đi qua Git + CD.
- Nâng cấp SOPS: đổi digest trong `deployments/scripts/sops-run.sh`, chạy
  `docker pull` trên VPS, deploy staging trước.

## 9. Troubleshooting

| Lỗi | Nguyên nhân / cách xử lý |
| --- | --- |
| `sops: command not found` | Chưa cài hoặc shell chưa nhận PATH; mở terminal mới, hoặc dùng `SOPS_MODE=docker` |
| `no matching creation rules found` | Chạy sai thư mục hoặc `path_regex` không khớp; chạy từ repo root, kiểm tra `.sops.yaml` |
| `no key could decrypt` | Identity không có trong recipient list hoặc sai `SOPS_AGE_KEY_FILE`; kiểm tra `sops decrypt` bằng key đã được cấp |
| `invalid character '#' looking for beginning of value` | Gọi `sops` trực tiếp trên file `.env.enc` mà thiếu `--input-type dotenv`; dùng `make env-verify`/`make env-edit` |
| `invalid dotenv input line` | File nguồn có CRLF hoặc dòng không có `=`; chuyển về LF trước khi encrypt |
| File `.enc` bị hỏng sau khi sửa | MAC không khớp; khôi phục từ Git rồi dùng `make env-edit` |
| `sops edit` mở editor lạ/không mở | Đặt `EDITOR`: `$env:EDITOR="code --wait"` hoặc `notepad` |
| Docker mode lỗi mount trên Windows | Dùng binary (`SOPS_MODE=binary`) trên máy dev |
| Decrypt được nhưng value có `\r` | File nguồn CRLF; tạo lại `.enc` từ file LF |
| `updatekeys` xong nhưng dev vẫn lỗi | Chưa commit/pull file `.enc` mới, hoặc dev dùng key khác |

## 10. Checklist

**Onboarding dev**

- [ ] Cài `sops` + `age`.
- [ ] Sinh key, backup private key.
- [ ] Gửi public key + handle + tên máy.
- [ ] Lead thêm vào `.sops.yaml`, `make env-updatekeys`, merge.
- [ ] Dev `git pull`, decrypt staging + production thành công.

**Offboarding**

- [ ] Xóa public key khỏi `.sops.yaml`.
- [ ] `make env-updatekeys`, merge.
- [ ] `sops rotate` cả hai file.
- [ ] Rotate provider/DB/Redis/RabbitMQ credential.
- [ ] Ghi nhận ngày rotate.

**Recovery drill (hằng năm)**

- [ ] Dùng recovery key decrypt production trên máy sạch.
- [ ] Xác nhận key VPS vẫn decrypt được sau drill.
- [ ] Ghi biên bản.
