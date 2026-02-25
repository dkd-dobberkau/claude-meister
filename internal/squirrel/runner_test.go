package squirrel

import (
	"testing"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner(30, "deep")
	if r.Days != 30 {
		t.Errorf("expected days=30, got %d", r.Days)
	}
	if r.Depth != "deep" {
		t.Errorf("expected depth=deep, got %s", r.Depth)
	}
}

func TestBuildArgs(t *testing.T) {
	r := NewRunner(14, "medium")
	args := r.buildArgs()
	expected := []string{"status", "--json", "--days", "14", "--depth", "medium"}
	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i, a := range args {
		if a != expected[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, expected[i], a)
		}
	}
}
