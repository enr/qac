//go:build !windows
// +build !windows

package qac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePath_EmptyDeclared_ReturnsBasedir(t *testing.T) {
	base := t.TempDir()
	result, err := resolvePath("", planContext{basedir: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != filepath.FromSlash(base) {
		t.Errorf("expected basedir %q, got %q", base, result)
	}
}

func TestResolvePath_RelativePath_JoinedWithBasedir(t *testing.T) {
	base := t.TempDir()
	result, err := resolvePath("subdir/file.txt", planContext{basedir: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected, _ := filepath.Abs(filepath.Join(base, "subdir", "file.txt"))
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestResolvePath_AbsolutePath_ReturnedAsIs(t *testing.T) {
	absPath := t.TempDir() // always absolute on all platforms
	// basedir should be ignored when path is absolute
	result, err := resolvePath(absPath, planContext{basedir: "/completely/different/base"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != absPath {
		t.Errorf("expected absolute path %q returned unchanged, got %q", absPath, result)
	}
}

func TestResolvePath_HomeTilde_ExpandsToHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("home dir unavailable:", err)
	}
	result, err := resolvePath("~/somefile", planContext{basedir: "/base"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.HasPrefix(result, "~") {
		t.Error("expected ~ to be expanded, got:", result)
	}
	if !strings.HasPrefix(result, home) {
		t.Errorf("expected result to start with home dir %q, got %q", home, result)
	}
}

func TestResolvePath_ParentDotDot_Resolved(t *testing.T) {
	base := t.TempDir()
	result, err := resolvePath("..", planContext{basedir: base})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected, _ := filepath.Abs(filepath.Join(base, ".."))
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
