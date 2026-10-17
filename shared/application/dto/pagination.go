package dto

import "encoding/base64"

type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
}

func EncodeCursor(value string) string { return base64.RawURLEncoding.EncodeToString([]byte(value)) }
func DecodeCursor(cursor string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	return string(raw), err
}

// AdminPage is reserved for administrative tables that need random access.
type AdminPage[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"totalItems"`
}
