package timeutils

import "time"

const BusinessTimezone = "Asia/Ho_Chi_Minh"

func BusinessLocation() *time.Location {
	loc, err := time.LoadLocation(BusinessTimezone)
	if err != nil {
		return time.FixedZone("ICT", 7*60*60)
	}
	return loc
}
func FormatAPI(t time.Time) string { return t.In(BusinessLocation()).Format(time.RFC3339) }
