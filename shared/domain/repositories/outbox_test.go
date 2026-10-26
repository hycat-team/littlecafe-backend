package repositories

import "testing"

func TestJSONFieldValue(t *testing.T) {
	var empty JSONField
	value, err := empty.Value()
	if err != nil || value != nil {
		t.Fatalf("empty Value = %v, %v", value, err)
	}

	value, err = JSONField(`{"a":1}`).Value()
	if err != nil || value != `{"a":1}` {
		t.Fatalf("Value = %v, %v", value, err)
	}
}

func TestJSONFieldScan(t *testing.T) {
	var field JSONField
	if err := field.Scan(nil); err != nil || field != nil {
		t.Fatalf("Scan(nil) = %v, %v", field, err)
	}
	if err := field.Scan([]byte(`{"a":1}`)); err != nil || string(field) != `{"a":1}` {
		t.Fatalf("Scan([]byte) = %s, %v", field, err)
	}
	if err := field.Scan(`{"a":2}`); err != nil || string(field) != `{"a":2}` {
		t.Fatalf("Scan(string) = %s, %v", field, err)
	}
	if err := field.Scan(42); err == nil {
		t.Fatal("unsupported type must fail")
	}
}

func TestOutboxRecordTableName(t *testing.T) {
	if got := (OutboxRecord{}).TableName(); got != "outbox_records" {
		t.Fatalf("TableName = %q", got)
	}
}
