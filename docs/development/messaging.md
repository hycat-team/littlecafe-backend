# Messaging

Exchange topic theo môi trường dùng tên `littlecafe.events`. Mỗi consumer có queue riêng; queue chính dùng retry có giới hạn và dead-letter sang `<queue>.dlq`.

Producer ghi event vào outbox trong cùng transaction với thay đổi local. Consumer ghi `messageId` vào inbox trước khi hoàn tất để xử lý at-least-once an toàn. Event chỉ được đăng ký sau khi AsyncAPI/JSON Schema xác định owner, version, payload và replay policy.

