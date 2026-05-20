package qac

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- Existence checks ---

func TestFileAssertion_ExistsTrue_FilePresent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "f.txt", "hello")
	r := (&FileAssertion{Path: "f.txt", Exists: true}).verify(planContext{basedir: dir})
	if !r.Success() {
		t.Errorf("expected success for existing file with Exists=true, got: %v", r.Errors())
	}
}

func TestFileAssertion_ExistsTrue_FileMissing(t *testing.T) {
	dir := t.TempDir()
	r := (&FileAssertion{Path: "missing.txt", Exists: true}).verify(planContext{basedir: dir})
	if r.Success() {
		t.Error("expected failure when file is absent but Exists=true")
	}
}

func TestFileAssertion_ExistsFalse_FileMissing(t *testing.T) {
	dir := t.TempDir()
	r := (&FileAssertion{Path: "missing.txt", Exists: false}).verify(planContext{basedir: dir})
	if !r.Success() {
		t.Errorf("expected success when file is absent and Exists=false, got: %v", r.Errors())
	}
}

func TestFileAssertion_ExistsFalse_FilePresent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "present.txt", "content")
	r := (&FileAssertion{Path: "present.txt", Exists: false}).verify(planContext{basedir: dir})
	if r.Success() {
		t.Error("expected failure when file is present but Exists=false")
	}
}

func TestFileAssertion_ExistsFalse_NoFurtherChecks(t *testing.T) {
	// When Exists=false and the file is indeed absent, no content checks run
	// and the assertion should succeed regardless of ContainsAll settings.
	dir := t.TempDir()
	r := (&FileAssertion{
		Path:        "missing.txt",
		Exists:      false,
		ContainsAll: []string{"unreachable"},
	}).verify(planContext{basedir: dir})
	if !r.Success() {
		t.Errorf("expected success: content checks should be skipped when Exists=false, got: %v", r.Errors())
	}
}

// --- ContainsAll ---

func TestFileAssertion_ContainsAll_AllPresent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "f.txt", "foo bar baz")
	r := (&FileAssertion{Path: "f.txt", Exists: true, ContainsAll: []string{"foo", "baz"}}).
		verify(planContext{basedir: dir})
	if !r.Success() {
		t.Errorf("expected success when all strings are in the file, got: %v", r.Errors())
	}
}

func TestFileAssertion_ContainsAll_OneMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "f.txt", "foo bar")
	r := (&FileAssertion{Path: "f.txt", Exists: true, ContainsAll: []string{"foo", "missing"}}).
		verify(planContext{basedir: dir})
	if r.Success() {
		t.Error("expected failure when file does not contain all required strings")
	}
}

// --- ContainsAny ---

func TestFileAssertion_ContainsAny_OnePresent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "f.txt", "foo bar")
	r := (&FileAssertion{Path: "f.txt", Exists: true, ContainsAny: []string{"missing", "foo"}}).
		verify(planContext{basedir: dir})
	if !r.Success() {
		t.Errorf("expected success when at least one string is in the file, got: %v", r.Errors())
	}
}

func TestFileAssertion_ContainsAny_NonePresent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "f.txt", "foo bar")
	r := (&FileAssertion{Path: "f.txt", Exists: true, ContainsAny: []string{"missing", "nope"}}).
		verify(planContext{basedir: dir})
	if r.Success() {
		t.Error("expected failure when file contains none of the required strings")
	}
}

// --- TextEqualsTo ---

func TestFileAssertion_TextEqualsTo_Pass(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "actual.txt", "line1\nline2\n")
	writeFile(t, dir, "expected.txt", "line1\nline2\n")
	r := (&FileAssertion{Path: "actual.txt", Exists: true, TextEqualsTo: "expected.txt"}).
		verify(planContext{basedir: dir})
	if !r.Success() {
		t.Errorf("expected success for identical text files, got: %v", r.Errors())
	}
}

func TestFileAssertion_TextEqualsTo_ContentDiffers(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "actual.txt", "line1\n")
	writeFile(t, dir, "expected.txt", "different\n")
	r := (&FileAssertion{Path: "actual.txt", Exists: true, TextEqualsTo: "expected.txt"}).
		verify(planContext{basedir: dir})
	if r.Success() {
		t.Error("expected failure when file content differs from expected")
	}
}

func TestFileAssertion_TextEqualsTo_ExpectedFileMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "actual.txt", "line1\n")
	r := (&FileAssertion{Path: "actual.txt", Exists: true, TextEqualsTo: "nonexistent.txt"}).
		verify(planContext{basedir: dir})
	if r.Success() {
		t.Error("expected failure when the expected file does not exist")
	}
}

// --- Binary detection ---

func TestIsBinaryFile_TextFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "text.txt", "hello world\nno null bytes here")
	f, err := os.Open(filepath.Join(dir, "text.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if isBinaryFile(f) {
		t.Error("expected text file to not be detected as binary")
	}
}

func TestIsBinaryFile_BinaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")
	if err := os.WriteFile(path, []byte{0x48, 0x65, 0x00, 0x6c, 0x6c, 0x6f}, 0644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if !isBinaryFile(f) {
		t.Error("expected file with null byte to be detected as binary")
	}
}

func TestIsBinaryFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "empty.txt", "")
	f, err := os.Open(filepath.Join(dir, "empty.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if isBinaryFile(f) {
		t.Error("expected empty file to not be detected as binary")
	}
}
