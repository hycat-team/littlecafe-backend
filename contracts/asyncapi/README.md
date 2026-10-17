# AsyncAPI và JSON Schema

Exchange chuẩn là `littlecafe.events`, kiểu topic. Mỗi consumer sở hữu queue riêng, retry có giới hạn và DLQ riêng. Envelope phải có `messageId`, `eventType`, `eventVersion`, `occurredAt`, `producer`, `traceId` và `payload`.

Bootstrap để trống channels. Chỉ thêm event sau khi owner, schema, idempotency và chiến lược replay được duyệt.

