package qac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- YAML parsing ---

func TestStdinFieldAccepted(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: cat
      stdin: "hello world"
    expectations:
      status:
        equals_to: 0
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Specs["test"].Command.Stdin != "hello world" {
		t.Errorf("stdin = %q, want %q", plan.Specs["test"].Command.Stdin, "hello world")
	}
}

func TestStdinFileFieldAccepted(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: cat
      stdin_file: ./testdata/input.txt
    expectations:
      status:
        equals_to: 0
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Specs["test"].Command.StdinFile != "./testdata/input.txt" {
		t.Errorf("stdin_file = %q, want %q", plan.Specs["test"].Command.StdinFile, "./testdata/input.txt")
	}
}

func TestStdinTypoRejected(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: cat
      stdinn: "hello"
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'stdinn', got nil")
	}
	if !strings.Contains(err.Error(), "stdinn") {
		t.Errorf("error should mention 'stdinn', got: %v", err)
	}
}

func TestStdinFileTypoRejected(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: cat
      stdin_fille: ./input.txt
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'stdin_fille', got nil")
	}
	if !strings.Contains(err.Error(), "stdin_fille") {
		t.Errorf("error should mention 'stdin_fille', got: %v", err)
	}
}

// --- Launcher integration: inline stdin ---

func TestLauncher_InlineStdin(t *testing.T) {
	zero := 0
	spec := Spec{
		Command: Command{
			Cli:   "cat",
			Stdin: "hello stdin\n",
		},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
			OutputAssertions: OutputAssertions{
				Stdout: OutputAssertion{EqualsTo: "hello stdin"},
			},
		},
	}
	sut := NewLauncher()
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
}

func TestLauncher_InlineStdinMultiLine(t *testing.T) {
	zero := 0
	spec := Spec{
		Command: Command{
			Cli:   "cat",
			Stdin: "line1\nline2\n",
		},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
			OutputAssertions: OutputAssertions{
				Stdout: OutputAssertion{ContainsAll: []string{"line1", "line2"}},
			},
		},
	}
	sut := NewLauncher()
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
}

// --- Launcher integration: stdin_file ---

func TestLauncher_StdinFile(t *testing.T) {
	dir := t.TempDir()
	inputFile := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(inputFile, []byte("from file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	zero := 0
	spec := Spec{
		Command: Command{
			Cli:       "cat",
			StdinFile: inputFile,
		},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
			OutputAssertions: OutputAssertions{
				Stdout: OutputAssertion{EqualsTo: "from file"},
			},
		},
	}
	sut := NewLauncher()
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
}

func TestLauncher_StdinFileMissing(t *testing.T) {
	spec := Spec{
		Command: Command{
			Cli:       "cat",
			StdinFile: "/nonexistent/path/input.txt",
		},
	}
	sut := NewLauncher()
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if len(res.AllErrors()) == 0 {
		t.Error("expected error for missing stdin_file, got none")
	}
}

// --- Mutual exclusion ---

func TestLauncher_StdinFileMissing_ErrorMentionsPath(t *testing.T) {
	const missingPath = "/nonexistent/qac_test_stdin_file_path.txt"
	spec := Spec{
		Command: Command{
			Cli:       "cat",
			StdinFile: missingPath,
		},
	}
	res := NewLauncher().Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	errs := res.AllErrors()
	if len(errs) == 0 {
		t.Fatal("expected error for missing stdin_file, got none")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "qac_test_stdin_file_path.txt") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("error message should mention the missing file path, got: %v", errs)
	}
}

func TestLauncher_StdinAndStdinFileMutuallyExclusive(t *testing.T) {
	spec := Spec{
		Command: Command{
			Cli:       "cat",
			Stdin:     "inline",
			StdinFile: "./some-file.txt",
		},
	}
	sut := NewLauncher()
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if len(res.AllErrors()) == 0 {
		t.Error("expected error when both stdin and stdin_file are set")
	}
}
