package config

import (
	"testing"
)

func TestLoad_AllVarsSet(t *testing.T) {
	t.Setenv("WORK_USER", "alice")
	t.Setenv("EDITOR", "nvim")
	t.Setenv("VISUAL", "code")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.User != "alice" {
		t.Errorf("user: got %q, want alice", cfg.User)
	}
	if cfg.Editor != "nvim" {
		t.Errorf("editor: got %q, want nvim", cfg.Editor)
	}
}

func TestLoad_EditorFallsBackToVisual(t *testing.T) {
	t.Setenv("WORK_USER", "bob")
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "code")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Editor != "code" {
		t.Errorf("editor: got %q, want code", cfg.Editor)
	}
}

func TestLoad_EditorFallsBackToVi(t *testing.T) {
	t.Setenv("WORK_USER", "bob")
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Editor != "vi" {
		t.Errorf("editor: got %q, want vi", cfg.Editor)
	}
}

func TestLoad_UserFallsBackToGitOrSystem(t *testing.T) {
	t.Setenv("WORK_USER", "")
	t.Setenv("EDITOR", "vi")
	t.Setenv("VISUAL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.User == "" {
		t.Error("user should not be empty")
	}
}

func TestLoad_EDITORTakesPrecedenceOverVISUAL(t *testing.T) {
	t.Setenv("WORK_USER", "alice")
	t.Setenv("EDITOR", "nano")
	t.Setenv("VISUAL", "code")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Editor != "nano" {
		t.Errorf("editor: got %q, want nano (EDITOR should take precedence)", cfg.Editor)
	}
}
