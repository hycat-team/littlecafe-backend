# Core Flows (service-level)

Tài liệu này mô tả luồng ở mức service/backend cho các workflow đã có trong
`littlecafe-docs/workflows/`. Nguồn nghiệp vụ và policy là các file đó; tài liệu
này chỉ ánh xạ sang ranh giới service, contract và invariant kỹ thuật. Không vẽ
diagram; khi nghiệp vụ và tài liệu này lệch nhau, báo conflict thay vì tự chọn.

Quy ước chung cho mọi luồng:

- Mutation quan trọng phải idempotent (`Idempotency-Key`), không vượt quyền,
  không truy cập database chéo service.
- Trạng thái tài chính/capacity không hoàn tất dựa trên một lần thử chưa xác minh.
- Gọi đồng bộ xuyên service qua gRPC; sự kiện bất đồng bộ qua RabbitMQ +
  transactional outbox. Saga do Commerce hoặc Experience sở hữu.

## WF-001 Account & Guardian-Child

- **Trạng thái:** Confirmed ở mức capability; policy mở.
- **Service:** Commerce — module Identity (account, auth, GuardianLink, RBAC).
- **Luồng backend:** đăng ký/đăng nhập → Guardian liên kết Child (`GuardianLink`)
  → kiểm tra RBAC + ownership + GuardianLink trước mọi dữ liệu trẻ em.
- **Contract dự kiến:** OpenAPI `/api/v1/identity/**`; event account (khi được
  duyệt).
- **Mở:** revoke, minimum data, retention của GuardianLink.

## WF-002 Drink ordering

- **Trạng thái:** Confirmed ở mức capability; policy mở.
- **Service:** Commerce — module Order và Wallet.
- **Luồng backend:** tạo order idempotent → Wallet debit **trong cùng
  transaction** (không saga) → staff nhận/chế biến/phục vụ → hoàn tất; huỷ trước
  chế biến hoặc không phục vụ được → credit hoàn toàn bộ.
- **Mở:** preorder timing, evidence, walk-in.

## WF-003 Activity booking

- **Trạng thái:** Confirmed ở mức capability; policy mở.
- **Service:** Experience — module Activity; Commerce — Identity (Child,
  GuardianLink) và Wallet (payment) qua gRPC.
- **Luồng backend:** chọn session shared/private cho Child → kiểm tra
  GuardianLink, eligibility, capacity → booking/payment (saga do Experience sở
  hữu) → check-in theo assignment → ghi completion.
- **Mở:** phí, hold, cancel/refund.

## WF-004 Competition management

- **Trạng thái:** Confirmed, trừ scoring exception.
- **Service:** Experience — module Activity; Commerce — Identity và Wallet
  (payment/waive) qua gRPC.
- **Luồng backend:** chọn Child → eligibility/capacity → hold tối đa 15 phút →
  payment hoặc waive hợp lệ → check-in → thi → ghi và publish kết quả theo quyền.
- **Mở:** scoring, tie, consent leaderboard, exception chi tiết.

## WF-005 Wallet & payment

- **Trạng thái:** Confirmed ở mức capability; quyền collection/payout cần xác minh.
- **Service:** Commerce — module Wallet; payOS là provider bên thứ ba.
- **Luồng backend:** top-up → payOS QR → verify webhook/signature → credit ledger;
  service yêu cầu payment → wallet debit idempotent qua gRPC → trả kết quả; event
  phát qua outbox cho Platform. Cash competition ghi riêng.
- **Mở:** quyền collection/payout thực tế.

## WF-006 Refund & withdrawal

- **Trạng thái:** Confirmed ở mức capability; policy mở.
- **Service:** service sở hữu xác định eligibility/amount; Commerce — module
  Wallet tạo refund obligation + credit; Manager duyệt withdrawal; payout hoặc
  manual fallback.
- **Luồng backend:** refund chỉ complete sau credit; withdrawal chỉ tạo sau khi
  refund credit xong → duyệt → payout (feature gate) hoặc manual → xác minh
  outcome.
- **Mở:** timeline/limit, retention bank snapshot.

## WF-007 3D guides

- **Trạng thái:** Confirmed ở mức capability; storage/device mở.
- **Service:** Experience — module Activity/Content (ContentVersion, asset);
  fallback media 2D qua Cloudinary.
- **Luồng backend:** mở game → tải ContentVersion có hiệu lực → tải asset 3D →
  hiển thị layout/objective/steps → fallback text/image/2D khi lỗi → quay lại
  lịch và đăng ký.
- **Mở:** storage asset 3D, target device, interaction.

## WF-008 RAG assistant

- **Trạng thái:** Confirmed ở mức capability; evaluation/freshness mở.
- **Service:** AI (retrieval, trả lời); Commerce (Identity authorize) và
  Experience (public content replicate) qua gRPC read-only.
- **Luồng backend:** nhận câu hỏi → xác định guest/user + intent → làm rõ →
  authorize → retrieve public content hoặc live private data → trả lời kèm
  nguồn/version → dẫn tới chức năng chính nếu cần mutation. RAG không thực hiện
  mutation.
- **Mở:** freshness, conflict handling, evaluation threshold. Xem `ai-service.md`
  cho mâu thuẫn provider.

## Chưa triển khai

Không workflow nào có endpoint/RPC/event/saga trong code hiện tại. Khi implement
một luồng, thứ tự bắt buộc: requirement/workflow/ADR → contract → migration +
entity → use case + test → event (nếu mutation liên service).
