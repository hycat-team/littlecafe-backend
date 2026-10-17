package timeutils

import "time"

// BusinessTimezone is the single business calendar for the backend.
const BusinessTimezone = "Asia/Ho_Chi_Minh"

// BusinessLocation returns the business timezone, falling back to a fixed
// UTC+7 zone when the IANA database is unavailable.
func BusinessLocation() *time.Location {
	loc, err := time.LoadLocation(BusinessTimezone)
	if err != nil {
		return time.FixedZone("ICT", 7*60*60)
	}
	return loc
}

// FormatAPI renders a timestamp as RFC3339 in the business timezone.
func FormatAPI(t time.Time) string { return t.In(BusinessLocation()).Format(time.RFC3339) }

// GetNow returns the current time in the given timezone; an empty timezone
// falls back to BusinessTimezone.
func GetNow(timezone string) time.Time {
	if timezone == "" {
		timezone = BusinessTimezone
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Now().In(BusinessLocation())
	}
	return time.Now().In(loc)
}
