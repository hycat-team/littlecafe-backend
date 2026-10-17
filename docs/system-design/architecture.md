# Kiến trúc Backend LittleCafe

Backend là Go workspace gồm 9 module: 7 service, `shared` và `contracts`. Bảy service là Identity, Order, Activity, Wallet, AI, Notification và Reporting; module Content nằm trong Activity service.

Mỗi service sở hữu database riêng trên cùng PostgreSQL instance. Service khác không đọc bảng trực tiếp. Read xuyên service dùng gRPC; mutation dùng RabbitMQ với transactional outbox, inbox/dedup, retry và DLQ. Order hoặc Activity sở hữu saga khởi tạo; Notification và Reporting phản ứng theo choreography.

Identity, Order, Activity và Wallet host gRPC trên port 9101-9104 ở local. HTTP dùng 8101-8107; Nginx dùng 8080. Production API là `https://api.littlecafe.kids/api/v1`.

Redis là hạ tầng không authoritative cho cache, quota/rate limit và SSE Pub/Sub. Key bắt buộc có namespace `littlecafe:<env>:<service>:<purpose>`.

AI là một service độc lập sẽ được thiết kế ở phase sau. Provider, model, vector
store và retrieval policy chưa được chốt, nên shared chưa chứa adapter hoặc
configuration AI. Media 2D dùng Cloudinary; nơi lưu asset 3D chưa chốt.

