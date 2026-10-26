package entities

import "testing"

func TestSoftDeleteEntity(t *testing.T) {
	entity := &SoftDeleteEntity{}
	entity.SoftDelete()
	if !entity.IsDeleted {
		t.Fatal("SoftDelete must mark the entity as deleted")
	}
	entity.Undo()
	if entity.IsDeleted {
		t.Fatal("Undo must clear the deleted flag")
	}
}
