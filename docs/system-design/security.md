# Bảo mật

- Email/password dùng Argon2id; Google OAuth/OIDC dùng callback trên `api.littlecafe.kids`.
- Access token dùng ES256. Chỉ Identity giữ private key; các service khác chỉ nhận public key.
- Refresh token là 32 byte ngẫu nhiên Base64URL; chỉ lưu SHA-256 digest. Feature Identity phải triển khai rotation, revoke và reuse detection trong cùng transaction.
- Service-to-service dùng mTLS. Certificate do private CA cấp, nằm ngoài Git và mount read-only.
- Khi gRPC thay mặt người dùng, JWT người dùng đi qua metadata riêng với identity của service.
- RBAC tĩnh kết hợp ownership, Guardian-Child và assignment policy.
- Dữ liệu ngân hàng chỉ tồn tại trong encrypted snapshot của withdrawal, dùng AES-GCM với withdrawal ID làm associated data.

