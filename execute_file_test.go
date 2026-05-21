package qac

import (
	"fmt"
	"strings"
	"testing"
)

// fakeT is a minimal testing.TB stub for unit-testing helpers that call
// t.Fatalf / t.Errorf. Fatalf does NOT call runtime.Goexit — it just records
// the failure, which is safe because every call site has an explicit return
// immediately after Fatalf.
type fakeT struct {
	testing.TB // satisfy the interface; unexported methods panic if reached
	fataled    bool
	failed     bool
	errors     []string
}

func (f *fakeT) Helper()                              {}
func (f *fakeT) Failed() bool                         { return f.failed }
func (f *fakeT) Fatalf(format string, args ...any)    { f.fataled = true; f.failed = true; f.errors = append(f.errors, fmt.Sprintf(format, args...)) }
func (f *fakeT) Errorf(format string, args ...any)    { f.failed = true; f.errors = append(f.errors, fmt.Sprintf(format, args...)) }

func TestExecuteFile_MissingFile_ReturnsError(t *testing.T) {
	_, err := NewLauncher().ExecuteFile("definitely_nonexistent_plan_xyz.yaml")
	if err == nil {
		t.Fatal("expected non-nil error for missing plan file")
	}
	if !strings.Contains(err.Error(), "definitely_nonexistent_plan_xyz.yaml") {
		t.Errorf("error should mention the file path, got: %v", err)
	}
}

func TestExecuteFile_InvalidYAML_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.yaml", "specs: [not: a: valid: yaml")
	absPath := dir + "/bad.yaml"
	_, err := NewLauncher().ExecuteFile(absPath)
	if err == nil {
		t.Fatal("expected non-nil error for invalid YAML")
	}
}

func TestExecuteFile_ValidFile_ReturnsReportNoError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "plan.yaml", `
specs:
  ok:
    command:
      cli: echo hi
    expectations:
      status:
        equals_to: 0
`)
	report, err := NewLauncher().ExecuteFile(dir + "/plan.yaml")
	if err != nil {
		t.Fatalf("unexpected error for valid plan: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestExecuteFile_SpecFailure_ReportNotNilErrorNil(t *testing.T) {
	// A spec that fails (wrong expected exit code) should produce a non-nil
	// report with errors inside, but ExecuteFile itself must return nil error.
	dir := t.TempDir()
	writeFile(t, dir, "plan.yaml", `
specs:
  fail:
    command:
      cli: echo hi
    expectations:
      status:
        equals_to: 99
`)
	report, err := NewLauncher().ExecuteFile(dir + "/plan.yaml")
	if err != nil {
		t.Fatalf("ExecuteFile must not return error for spec-level failures, got: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.AllErrors()) == 0 {
		t.Error("expected spec failure to be recorded in report")
	}
}

// --- ExecuteFileT ---

func TestExecuteFileT_MissingFile_FatalsT(t *testing.T) {
	inner := &fakeT{}
	NewLauncher().ExecuteFileT(inner, "definitely_nonexistent_plan_xyz.yaml")
	if !inner.fataled {
		t.Error("ExecuteFileT should have called Fatalf for a missing plan file")
	}
}

func TestExecuteFileT_ValidFile_PassingPlan_NoFailure(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "plan.yaml", `
specs:
  ok:
    command:
      cli: echo hi
    expectations:
      status:
        equals_to: 0
`)
	inner := &fakeT{}
	report := NewLauncher().ExecuteFileT(inner, dir+"/plan.yaml")
	if inner.failed {
		t.Errorf("ExecuteFileT must not fail t when the plan passes; errors: %v", inner.errors)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestExecuteFileT_SpecFailure_ErrorfCalled(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "plan.yaml", `
specs:
  fail:
    command:
      cli: echo hi
    expectations:
      status:
        equals_to: 99
`)
	inner := &fakeT{}
	report := NewLauncher().ExecuteFileT(inner, dir+"/plan.yaml")
	if !inner.failed {
		t.Error("ExecuteFileT should have called Errorf for a spec-level failure")
	}
	if inner.fataled {
		t.Error("ExecuteFileT must not call Fatalf for spec-level failures")
	}
	if report == nil {
		t.Fatal("expected non-nil report even on spec failure")
	}
}

// --- ExecuteT ---

func TestExecuteT_PassingPlan_NoFailure(t *testing.T) {
	plan := NewPlan().
		Spec("ok", NewSpec().
			Command(ShellCmd("echo hi")).
			ExpectStatus(0)).
		Build()
	inner := &fakeT{}
	report := NewLauncher().ExecuteT(inner, plan)
	if inner.failed {
		t.Errorf("ExecuteT must not fail t when the plan passes; errors: %v", inner.errors)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestExecuteT_FailingPlan_ErrorfCalled(t *testing.T) {
	plan := NewPlan().
		Spec("fail", NewSpec().
			Command(ShellCmd("echo hi")).
			ExpectStatus(99)).
		Build()
	inner := &fakeT{}
	NewLauncher().ExecuteT(inner, plan)
	if !inner.failed {
		t.Error("ExecuteT should have called Errorf for a spec-level failure")
	}
}
