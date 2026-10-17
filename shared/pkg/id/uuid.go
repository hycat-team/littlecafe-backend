package id

import "github.com/google/uuid"

func New() (uuid.UUID, error) { return uuid.NewV7() }
func MustNew() uuid.UUID {
	value, err := New()
	if err != nil {
		panic(err)
	}
	return value
}
