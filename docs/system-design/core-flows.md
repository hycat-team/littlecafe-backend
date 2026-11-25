# Core Flows (service-level)

Tài liệu này mô tả luồng ở mức service/backend cho các workflow đã có trong
`littlecafe-docs/workflows/`. Nguồn nghiệp vụ và policy là các file đó; tài liệu
này chỉ ánh xạ sang ranh giới service, contract và invariant kỹ thuật. Không vẽ
diagram; khi nghiệp vụ và tài liệu này lệch nhau, báo conflict thay vì tự chọn.

Quy ước chung cho mỗi luồng:

- Mutation quan trọng phải idempotent (`Idempotency-Key`), không vượt quyền,
  không truy cập database chéo service.
- Role + branch + ownership/GuardianLink/assignment được kiểm tra trước mutation;
  cross-branch trả `403`.
- Trạng thái tài chính/capacity không hoàn tất dựa trên một lần thử chưa xác minh.
- Gọi đồng bộ xuyên service qua gRPC; sự kiện bất đồng bộ qua RabbitMQ +
  transactional outbox. Saga do Commerce hoặc Experience sở hữu.
- 4 role `GUARDIAN`/`STAFF`/`MANAGER`/`ADMIN`; multi-role/multi-branch active grant;
  `STAFF`/`MANAGER` bắt buộc `branchId` (DEC-013).

## WF-001 Account & Guardian-Child

- **Trạng thái:** Confirmed.
- **Service:** Commerce/Identity (account, auth, AccountRoleGrant, Branch, GuardianLink, RBAC).
- **Luồng backend:** đăng ký/đăng nhập → Guardian liên kết Child (`GuardianLink`) →
  Admin/Manager grant role theo branch → kiểm tra role + branch + ownership +
  GuardianLink trước mọi dữ liệu trẻ em.
- **Contract dự kiến:** OpenAPI `/api/v1/identity/**`; event account/role (khi được duyệt).
- **Mở:** không có OQ mới; role grant history/audit theo DEC-013.

## WF-002 Drink ordering

- **Trạng thái:** Confirmed.
- **Service:** Commerce/Order và Wallet.
- **Luồng backend:** tạo order idempotent, chốt price snapshot, gắn `branchId` →
  payment (wallet/direct QR/cash) → Staff nhận/chế biến/phục vụ → hoàn tất; hủy
  trước prep hoặc không phục vụ được → credit hoàn toàn bộ.
- **Walk-in:** Staff tạo order tại counter, QR/cash, transaction only; không ví/pre-order.
- **Mở:** không có OQ mới; hoàn một phần sau prep chưa chốt.

## WF-003 Activity booking

- **Trạng thái:** Confirmed.
- **Service:** Experience/Activity; Commerce/Identity (Child, GuardianLink) và Wallet
  (payment) qua gRPC.
- **Luồng backend:** chọn session shared/private cho Child → kiểm tra GuardianLink,
  eligibility, capacity (atomic) → booking/payment (không hold) → check-in/complete
  theo assignment; walk-in regular activity do Staff tạo record tại counter.
- **Mở:** không có OQ mới; hủy sau giờ bắt đầu là ngoại lệ vận hành.

## WF-004 Competition management

- **Trạng thái:** Confirmed.
- **Service:** Experience/Activity; Commerce/Identity và Wallet (payment/waive) qua gRPC.
- **Luồng backend:** chọn Child → eligibility/capacity → hold tối đa 15 phút →
  payment/waive → check-in → thi → ghi/publish kết quả theo quyền.
- **Walk-in:** từ chối; competition yêu cầu Guardian đăng nhập + GuardianLink.
- **Mở:** tie-breaker mặc định điểm + thời gian; policy chi tiết theo competition.

## WF-005 Wallet & payment

- **Trạng thái:** Confirmed; payout verification gate.
- **Service:** Commerce/Wallet; payOS là provider.
- **Luồng backend:** top-up → payOS QR → verify webhook/signature → credit ledger;
  service yêu cầu payment → wallet debit idempotent qua gRPC → trả kết quả; direct
  QR không credit ví; cash receipt tại counter không tạo top-up; event qua outbox.
- **Mở:** payOS payout/collection capability verification; feature flag off.

## WF-006 Refund & withdrawal

- **Trạng thái:** Confirmed.
- **Service:** service sở hữu xác định eligibility/amount; Commerce/Wallet tạo
  refund obligation + credit; Manager duyệt withdrawal; payout/manual fallback.
- **Luồng backend:** refund chỉ complete sau credit; withdrawal chỉ từ available
  balance → duyệt → payout (feature gate) hoặc manual → xác minh outcome.
- **Mở:** không có OQ mới; payout uncertain giữ pending.

## WF-007 3D guides

- **Trạng thái:** Confirmed.
- **Service:** Experience/Content (ContentVersion, guide); GLB static trên Nginx/CDN;
  media 2D/fallback Cloudinary.
- **Luồng backend:** mở game → tải published ContentVersion → tải GLB/2D →
  hiển thị layout/objective/steps → fallback text/image/2D khi lỗi → quay lại lịch.
- **Mở:** không có OQ mới; target devices/loading theo DEC-019.

## WF-008 RAG assistant

- **Trạng thái:** Boundary Confirmed; provider/model/vector store/retrieval policy Undecided (OQ-012).
- **Service:** AI (retrieval/answer); Commerce/Identity (authorize) và Experience
  (public/operational content) qua gRPC read-only.
- **Luồng backend:** nhận câu hỏi → xác định guest/user + intent → làm rõ →
  authorize → retrieve public/private → trả lời kèm nguồn/version → dẫn tới
  chức năng chính nếu cần mutation. RAG không mutation.
- **Mở:** provider/model/vector store/retrieval policy; không tạo adapter/contract
  AI đến khi chốt.

## Chưa triển khai

Không workflow nào có endpoint/RPC/event/saga trong code hiện tại. Khi implement
một luồng, thứ tự bắt buộc: requirement/workflow/ADR → contract → migration +
entity → use case + test → event (nếu mutation liên service).
