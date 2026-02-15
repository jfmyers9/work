package editor

import (
	"testing"
)

func TestResolveEditor_EDITORTakesPrecedence(t *testing.T) {
	t.Setenv("EDITOR", "nano")
	t.Setenv("VISUAL", "code")

	if got := ResolveEditor(); got != "nano" {
		t.Errorf("got %q, want %q", got, "nano")
	}
}

func TestResolveEditor_FallsBackToVISUAL(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "code")

	if got := ResolveEditor(); got != "code" {
		t.Errorf("got %q, want %q", got, "code")
	}
}

func TestResolveEditor_FallsBackToVi(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")

	if got := ResolveEditor(); got != "vi" {
		t.Errorf("got %q, want %q", got, "vi")
	}
}
