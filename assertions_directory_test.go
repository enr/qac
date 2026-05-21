package qac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkDir(t *testing.T, parent, name string) string {
	t.Helper()
	p := filepath.Join(parent, name)
	if err := os.Mkdir(p, 0755); err != nil {
		t.Fatal(err)
	}
	return p
}

// --- Existence ---

func TestDirectoryAssertion_ExistsTrue_DirPresent(t *testing.T) {
	base := t.TempDir()
	mkDir(t, base, "mydir")
	r := (&DirectoryAssertion{Path: "mydir", Exists: boolPtr(true)}).verify(planContext{basedir: base})
	if !r.Success() {
		t.Errorf("expected success for existing dir with Exists=true, got: %v", r.Errors())
	}
}

func TestDirectoryAssertion_ExistsTrue_DirMissing(t *testing.T) {
	base := t.TempDir()
	r := (&DirectoryAssertion{Path: "absent", Exists: boolPtr(true)}).verify(planContext{basedir: base})
	if r.Success() {
		t.Error("expected failure when directory does not exist but Exists=true")
	}
}

func TestDirectoryAssertion_ExistsFalse_DirMissing(t *testing.T) {
	base := t.TempDir()
	r := (&DirectoryAssertion{Path: "absent", Exists: boolPtr(false)}).verify(planContext{basedir: base})
	if !r.Success() {
		t.Errorf("expected success when directory absent and Exists=false, got: %v", r.Errors())
	}
}

func TestDirectoryAssertion_ExistsFalse_DirPresent(t *testing.T) {
	base := t.TempDir()
	mkDir(t, base, "present")
	r := (&DirectoryAssertion{Path: "present", Exists: boolPtr(false)}).verify(planContext{basedir: base})
	if r.Success() {
		t.Error("expected failure when directory exists but Exists=false")
	}
}

func TestDirectoryAssertion_ExistsFalse_NoFurtherChecks(t *testing.T) {
	// Content checks must not run when the directory is confirmed absent.
	base := t.TempDir()
	r := (&DirectoryAssertion{
		Path:        "absent",
		Exists:      boolPtr(false),
		ContainsAll: []string{"should-not-be-checked"},
	}).verify(planContext{basedir: base})
	if !r.Success() {
		t.Errorf("expected success: content checks skipped when Exists=false, got: %v", r.Errors())
	}
}

func TestDirectoryAssertion_PathIsFile(t *testing.T) {
	base := t.TempDir()
	writeFile(t, base, "notadir.txt", "content")
	r := (&DirectoryAssertion{Path: "notadir.txt", Exists: boolPtr(true)}).verify(planContext{basedir: base})
	if r.Success() {
		t.Error("expected failure when path points to a file, not a directory")
	}
}

// --- ContainsAll (this also validates the value-receiver bug was fixed) ---

func TestDirectoryAssertion_ContainsAll_AllPresent(t *testing.T) {
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "a.txt", "")
	writeFile(t, d, "b.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsAll: []string{"a.txt", "b.txt"}}).
		verify(planContext{basedir: base})
	if !r.Success() {
		t.Errorf("expected success for contains_all, got: %v", r.Errors())
	}
}

func TestDirectoryAssertion_ContainsAll_OneMissing(t *testing.T) {
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "a.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsAll: []string{"a.txt", "missing.txt"}}).
		verify(planContext{basedir: base})
	if r.Success() {
		t.Error("expected failure when directory is missing a required file")
	}
}

// --- ContainsAny ---

func TestDirectoryAssertion_ContainsAny_OnePresent(t *testing.T) {
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "a.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsAny: []string{"missing.txt", "a.txt"}}).
		verify(planContext{basedir: base})
	if !r.Success() {
		t.Errorf("expected success for contains_any, got: %v", r.Errors())
	}
}

func TestDirectoryAssertion_ContainsAny_NonePresent(t *testing.T) {
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "a.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsAny: []string{"x.txt", "y.txt"}}).
		verify(planContext{basedir: base})
	if r.Success() {
		t.Error("expected failure when directory contains none of the required files")
	}
}

func TestDirectoryAssertion_ContainsAny_ErrorShowsBothLists(t *testing.T) {
	// The error message must include both what was expected (ContainsAny) and
	// what is actually present, so the user can compare without manual inspection.
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "present.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsAny: []string{"wanted.txt"}}).
		verify(planContext{basedir: base})
	if r.Success() {
		t.Fatal("expected failure")
	}
	errs := r.Errors()
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	msg := errs[0].Error()
	if !strings.Contains(msg, "wanted.txt") {
		t.Errorf("error should mention the expected file 'wanted.txt', got: %q", msg)
	}
	if !strings.Contains(msg, "present.txt") {
		t.Errorf("error should mention the actual file 'present.txt' for debugging, got: %q", msg)
	}
}

// --- ContainsExactly (also validates the value-receiver bug was fixed) ---

func TestDirectoryAssertion_ContainsExactly_Match(t *testing.T) {
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "a.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsExactly: []string{"a.txt"}}).
		verify(planContext{basedir: base})
	if !r.Success() {
		t.Errorf("expected success for exact directory contents match, got: %v", r.Errors())
	}
}

func TestDirectoryAssertion_ContainsExactly_ExtraFile(t *testing.T) {
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "a.txt", "")
	writeFile(t, d, "b.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsExactly: []string{"a.txt"}}).
		verify(planContext{basedir: base})
	if r.Success() {
		t.Error("expected failure when directory has more files than expected")
	}
}

func TestDirectoryAssertion_ContainsExactly_MissingFile(t *testing.T) {
	base := t.TempDir()
	d := mkDir(t, base, "d")
	writeFile(t, d, "a.txt", "")
	r := (&DirectoryAssertion{Path: "d", Exists: boolPtr(true), ContainsExactly: []string{"a.txt", "b.txt"}}).
		verify(planContext{basedir: base})
	if r.Success() {
		t.Error("expected failure when directory is missing a file listed in contains_exactly")
	}
}
