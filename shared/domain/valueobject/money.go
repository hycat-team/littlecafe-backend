package valueobject

import "fmt"

// MoneyVND stores whole Vietnamese dong. Floating-point money is forbidden.
type MoneyVND int64

func NewMoneyVND(amount int64) (MoneyVND, error) {
	if amount < 0 {
		return 0, fmt.Errorf("money: amount must be non-negative")
	}
	return MoneyVND(amount), nil
}
func (m MoneyVND) Int64() int64 { return int64(m) }
