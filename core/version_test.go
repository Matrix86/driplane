package core

import "testing"

func TestVersionConstants(t *testing.T) {
	if Name == "" {
		t.Error("Name constant should not be empty")
	}
	if Name != "driplane" {
		t.Errorf("expected Name 'driplane', got '%s'", Name)
	}

	if Version == "" {
		t.Error("Version constant should not be empty")
	}

	if Author == "" {
		t.Error("Author constant should not be empty")
	}
}
