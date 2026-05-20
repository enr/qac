package qac

import (
	"os"
	"path/filepath"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func outputCtx(stdout, stderr string) planContext {
	return planContext{
		commandResult: executionResult{stdout: stdout, stderr: stderr},
	}
}

// --- IsEmpty ---

func TestOutputIsEmpty_ShouldBeEmptyAndIs(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", IsEmpty: boolPtr(true)}).verify(outputCtx("", ""))
	if !r.Success() {
		t.Errorf("expected success for empty output with is_empty=true, got: %v", r.Errors())
	}
}

func TestOutputIsEmpty_ShouldBeEmptyButIsNot(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", IsEmpty: boolPtr(true)}).verify(outputCtx("not empty", ""))
	if r.Success() {
		t.Error("expected failure when output is not empty but is_empty=true")
	}
}

func TestOutputIsEmpty_NilIsNeverChecked(t *testing.T) {
	// IsEmpty nil means the check is skipped entirely.
	r := (&OutputAssertion{id: "stdout"}).verify(outputCtx("anything", ""))
	if !r.Success() {
		t.Errorf("expected success when IsEmpty is nil, got: %v", r.Errors())
	}
}

// --- StartsWith / EndsWith ---

func TestOutputStartsWith_Pass(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", StartsWith: "hello"}).verify(outputCtx("hello world", ""))
	if !r.Success() {
		t.Errorf("expected success for starts_with, got: %v", r.Errors())
	}
}

func TestOutputStartsWith_Fail(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", StartsWith: "world"}).verify(outputCtx("hello world", ""))
	if r.Success() {
		t.Error("expected failure when output does not start with expected prefix")
	}
}

func TestOutputEndsWith_Pass(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", EndsWith: "world"}).verify(outputCtx("hello world", ""))
	if !r.Success() {
		t.Errorf("expected success for ends_with, got: %v", r.Errors())
	}
}

func TestOutputEndsWith_Fail(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", EndsWith: "hello"}).verify(outputCtx("hello world", ""))
	if r.Success() {
		t.Error("expected failure when output does not end with expected suffix")
	}
}

// --- ContainsAll ---

func TestOutputContainsAll_AllPresent(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsAll: []string{"foo", "bar", "baz"}}).
		verify(outputCtx("foo bar baz", ""))
	if !r.Success() {
		t.Errorf("expected success when all strings are present, got: %v", r.Errors())
	}
}

func TestOutputContainsAll_OneMissing(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsAll: []string{"foo", "missing"}}).
		verify(outputCtx("foo bar", ""))
	if r.Success() {
		t.Error("expected failure when one of the required strings is absent")
	}
}

// --- ContainsAny ---

func TestOutputContainsAny_OnePresent(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsAny: []string{"missing", "world"}}).
		verify(outputCtx("hello world", ""))
	if !r.Success() {
		t.Errorf("expected success when at least one string is present, got: %v", r.Errors())
	}
}

func TestOutputContainsAny_NonePresent(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsAny: []string{"missing", "nope"}}).
		verify(outputCtx("hello world", ""))
	if r.Success() {
		t.Error("expected failure when none of the strings are present")
	}
}

// --- Stderr vs stdout selection ---

func TestOutputStderr_ReadsStderr(t *testing.T) {
	r := (&OutputAssertion{id: "stderr", EqualsTo: "err-value"}).
		verify(outputCtx("out-value", "err-value"))
	if !r.Success() {
		t.Errorf("expected success reading stderr, got: %v", r.Errors())
	}
}

func TestOutputStderr_DoesNotReadStdout(t *testing.T) {
	r := (&OutputAssertion{id: "stderr", EqualsTo: "out-value"}).
		verify(outputCtx("out-value", "err-value"))
	if r.Success() {
		t.Error("stderr assertion matched stdout content; it should read stderr only")
	}
}

// --- Leading/trailing whitespace trimming ---

func TestOutputTrimming(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", EqualsTo: "hello"}).
		verify(outputCtx("  hello  \n", ""))
	if !r.Success() {
		t.Errorf("expected output to be trimmed before comparison, got: %v", r.Errors())
	}
}

// --- EqualsToFile ---

func TestOutputEqualsToFile_Pass(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "expected.txt"), []byte("expected content"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := planContext{
		basedir:       dir,
		commandResult: executionResult{stdout: "expected content"},
	}
	r := (&OutputAssertion{id: "stdout", EqualsToFile: "expected.txt"}).verify(ctx)
	if !r.Success() {
		t.Errorf("expected success for equals_to_file, got: %v", r.Errors())
	}
}

func TestOutputEqualsToFile_ContentMismatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "expected.txt"), []byte("expected content"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := planContext{
		basedir:       dir,
		commandResult: executionResult{stdout: "different content"},
	}
	r := (&OutputAssertion{id: "stdout", EqualsToFile: "expected.txt"}).verify(ctx)
	if r.Success() {
		t.Error("expected failure when output differs from file content")
	}
}

func TestOutputEqualsToFile_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	ctx := planContext{
		basedir:       dir,
		commandResult: executionResult{stdout: "some output"},
	}
	r := (&OutputAssertion{id: "stdout", EqualsToFile: "nonexistent.txt"}).verify(ctx)
	if r.Success() {
		t.Error("expected failure when the reference file does not exist")
	}
}
