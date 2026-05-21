package qac

import (
	"strings"
	"testing"
)

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
