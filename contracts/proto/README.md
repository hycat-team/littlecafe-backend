# Protobuf

Chỉ định nghĩa read-only cross-service contract. Identity, Order, Activity và Wallet là bốn gRPC host ban đầu. Không truyền entity nội bộ; message chỉ gồm primitive, UUID và enum đã version hóa. mTLS xác thực service, user JWT được truyền riêng khi hành động thay người dùng.

