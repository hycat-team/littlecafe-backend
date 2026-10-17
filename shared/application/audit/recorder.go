// Package audit defines the admin audit recording port.
package audit

import "context"

// Entry is one audit record.
type Entry struct {
	ActorID    string
	Action     string
	TargetType string
	TargetID   string
	Detail     map[string]any
}

// Recorder persists audit entries.
type Recorder interface {
	Record(ctx context.Context, entry Entry) error
}
