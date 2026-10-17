# Module wallet

Module này tuân theo Clean Architecture với bốn lớp `domain`, `application`, `infrastructure` và `presentation`.

Chưa có entity, use case, endpoint, RPC hoặc event nghiệp vụ. Chỉ bổ sung chúng sau khi workflow, quyền sở hữu dữ liệu và contract tương ứng được duyệt. Module `content` thuộc cùng service triển khai với `activity`, nhưng vẫn giữ ranh giới nghiệp vụ riêng.

