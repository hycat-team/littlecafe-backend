# Deferred — {{feature name}}

Công việc phát hiện khi làm {{feature name}} nhưng **không thuộc feature này**.
Không mục nào ở đây chặn việc đóng {{feature name}} và không mục nào là task của
{{feature name}}.

This file exists because an unchecked box in `tasks.md` makes a different claim.
`- [ ]` means *this feature is unfinished*. Work belonging to another feature, or
to the test infrastructure itself, is not this feature's unfinished business, and
recording it as such either keeps a closed feature looking open or pressures the
exit condition into being falsified.

## Quy tắc

- Mỗi entry phải nêu **vấn đề**, **vì sao ngoài phạm vi** và **điều kiện gỡ**.
- Entry liên quan task đã tick phải nêu task ID đó. Nếu finding làm mất
  acceptance evidence, reopen task và xử lý như in-scope/blocking; chỉ defer
  giới hạn không làm sai completion claim.
- Nếu công việc thực ra thuộc feature, chuyển nó về `tasks.md` thành task thật.
  **Đây không phải bãi đỗ cho in-scope work.**
- Báo mọi entry trong báo cáo cuối; không defer im lặng.

## Entries

<!--
Format each entry as:

### D-nn — <short title>

- **Vấn đề**: <mô tả đủ cụ thể để tái hiện hoặc đánh giá>
- **Vì sao ngoài phạm vi**: <ranh giới thật, không dùng "khó" hoặc "hết giờ">
- **Điều kiện gỡ**: <bước cụ thể và nơi công việc nên được xử lý>
- **Ảnh hưởng task đã tick**: <task IDs hoặc "không">

Delete this comment block once the first real entry exists.
-->
