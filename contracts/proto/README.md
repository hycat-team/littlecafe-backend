# Protobuf

Định nghĩa contract gRPC đồng bộ giữa các service: read và command idempotent. Commerce, Experience, AI và Platform là bốn gRPC host ban đầu. Không truyền entity nội bộ; message chỉ gồm primitive, UUID và enum đã version hóa. mTLS xác thực service, user JWT được truyền riêng khi hành động thay người dùng. Package dự kiến: `littlecafe.<service>.v1`.

