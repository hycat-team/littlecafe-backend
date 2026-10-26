package sse

import (
	"encoding/json"
	"testing"
	"time"
)

func testEvent(id string) Event {
	return Event{ID: id, Name: "updated", Data: json.RawMessage(`{"id":"` + id + `"}`)}
}

func drain(ch chan Event) []Event {
	var events []Event
	for {
		select {
		case event := <-ch:
			events = append(events, event)
		default:
			return events
		}
	}
}

func TestSubscribeReplaysAfterLastEventID(t *testing.T) {
	m := NewSSEManager(nil)
	m.Publish("stream", testEvent("1"))
	m.Publish("stream", testEvent("2"))
	m.Publish("stream", testEvent("3"))

	ch := m.Subscribe("stream", "1")
	defer m.Unsubscribe("stream", ch)

	got := drain(ch)
	if len(got) != 2 || got[0].ID != "2" || got[1].ID != "3" {
		t.Fatalf("replay = %+v, want ids 2,3", got)
	}
}

func TestSubscribeUnknownLastEventIDReplaysWholeBuffer(t *testing.T) {
	m := NewSSEManager(nil)
	m.Publish("stream", testEvent("1"))
	m.Publish("stream", testEvent("2"))

	ch := m.Subscribe("stream", "missing")
	defer m.Unsubscribe("stream", ch)

	got := drain(ch)
	if len(got) != 2 {
		t.Fatalf("replay = %+v, want two events", got)
	}
}

func TestSubscribeWithoutLastEventIDReplaysWholeBuffer(t *testing.T) {
	m := NewSSEManager(nil)
	m.Publish("stream", testEvent("1"))

	ch := m.Subscribe("stream", "")
	defer m.Unsubscribe("stream", ch)

	if got := drain(ch); len(got) != 1 {
		t.Fatalf("replay = %+v, want one event", got)
	}
}

func TestSubscribeSkipsExpiredBuffer(t *testing.T) {
	m := NewSSEManager(nil)
	base := time.Unix(1_760_000_000, 0)
	m.now = func() time.Time { return base }
	m.Publish("stream", testEvent("1"))
	m.now = func() time.Time { return base.Add(10 * time.Minute) }

	ch := m.Subscribe("stream", "")
	defer m.Unsubscribe("stream", ch)

	if got := drain(ch); len(got) != 0 {
		t.Fatalf("replay = %+v, want no events after TTL", got)
	}
}
