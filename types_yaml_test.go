package qac

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func unmarshalPlan(t *testing.T, src string) (*TestPlan, error) {
	t.Helper()
	plan := &TestPlan{}
	err := yaml.Unmarshal([]byte(src), plan)
	return plan, err
}

// Case 1: typo in command field name
func TestUnknownFieldInCommand(t *testing.T) {
	input := `
specs:
  test:
    commnad:
      cli: my-tool --output result.txt
    expectations:
      status:
        equals_to: 0
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'commnad', got nil")
	}
	if !strings.Contains(err.Error(), "commnad") {
		t.Errorf("error message should mention the unknown field 'commnad', got: %v", err)
	}
}

// Case 2: typo in assertion field name
func TestUnknownFieldInStatusAssertion(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      status:
        equal_to: 0
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'equal_to', got nil")
	}
	if !strings.Contains(err.Error(), "equal_to") {
		t.Errorf("error message should mention the unknown field 'equal_to', got: %v", err)
	}
}

// Case 3: text_equals_to used on a directory assertion
func TestTextEqualsToOnDirectoryAssertion(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      fs:
        - directory: ./output
          text_equals_to: ./expected
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	fa := plan.Specs["test"].Expectations.FileSystemAssertions[0]
	ctx := planContext{}
	_, actualErr := fa.actualAssertion(ctx)
	if actualErr == nil {
		t.Fatal("expected error for text_equals_to on directory assertion, got nil")
	}
	if !strings.Contains(actualErr.Error(), "text_equals_to") {
		t.Errorf("error should mention 'text_equals_to', got: %v", actualErr)
	}
}

// Case 3b: contains_exactly used on a file assertion
func TestContainsExactlyOnFileAssertion(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      fs:
        - file: ./output.txt
          contains_exactly:
            - line1
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	fa := plan.Specs["test"].Expectations.FileSystemAssertions[0]
	ctx := planContext{}
	_, actualErr := fa.actualAssertion(ctx)
	if actualErr == nil {
		t.Fatal("expected error for contains_exactly on file assertion, got nil")
	}
	if !strings.Contains(actualErr.Error(), "contains_exactly") {
		t.Errorf("error should mention 'contains_exactly', got: %v", actualErr)
	}
}

// Unknown field at plan level
func TestUnknownFieldAtPlanLevel(t *testing.T) {
	input := `
speks:
  test:
    command:
      cli: my-tool
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'speks' at plan level, got nil")
	}
	if !strings.Contains(err.Error(), "speks") {
		t.Errorf("error message should mention the unknown field 'speks', got: %v", err)
	}
}

func TestTimeoutFieldAccepted(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool process
      timeout: 30s
    expectations:
      status:
        equals_to: 0
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error parsing timeout field: %v", err)
	}
	if plan.Specs["test"].Command.Timeout != "30s" {
		t.Errorf("expected timeout %q, got %q", "30s", plan.Specs["test"].Command.Timeout)
	}
}

func TestTimeoutTypoRejected(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool process
      timeot: 30s
    expectations:
      status:
        equals_to: 0
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'timeot', got nil")
	}
	if !strings.Contains(err.Error(), "timeot") {
		t.Errorf("error message should mention the unknown field 'timeot', got: %v", err)
	}
}

func TestVarsSectionParsed(t *testing.T) {
	input := `
vars:
  base: /tmp/workdir
  tool: ./bin/mytool
specs:
  test:
    command:
      cli: mytool
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Vars["base"] != "/tmp/workdir" {
		t.Errorf("vars[base] = %q, want %q", plan.Vars["base"], "/tmp/workdir")
	}
	if plan.Vars["tool"] != "./bin/mytool" {
		t.Errorf("vars[tool] = %q, want %q", plan.Vars["tool"], "./bin/mytool")
	}
}

func TestVarsTypoRejected(t *testing.T) {
	input := `
varz:
  base: /tmp/workdir
specs:
  test:
    command:
      cli: mytool
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'varz' at plan level, got nil")
	}
	if !strings.Contains(err.Error(), "varz") {
		t.Errorf("error message should mention 'varz', got: %v", err)
	}
}

func TestRegexFieldsAccepted(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      output:
        stdout:
          matches: "^ok$"
          not_matches: "ERROR|WARN"
      fs:
        - file: out.log
          contains_matching: "duration: [0-9]+ms"
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stdout := plan.Specs["test"].Expectations.OutputAssertions.Stdout
	if stdout.Matches != "^ok$" {
		t.Errorf("matches = %q, want %q", stdout.Matches, "^ok$")
	}
	if stdout.NotMatches != "ERROR|WARN" {
		t.Errorf("not_matches = %q, want %q", stdout.NotMatches, "ERROR|WARN")
	}
	fs := plan.Specs["test"].Expectations.FileSystemAssertions
	if len(fs) != 1 || fs[0].ContainsMatching != "duration: [0-9]+ms" {
		t.Errorf("contains_matching not parsed correctly, got: %+v", fs)
	}
}

func TestRegexTyposRejected(t *testing.T) {
	cases := []struct {
		name  string
		field string
		input string
	}{
		{
			name:  "matchez typo",
			field: "matchez",
			input: `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      output:
        stdout:
          matchez: "^ok$"
`,
		},
		{
			name:  "contains_matchin typo",
			field: "contains_matchin",
			input: `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      fs:
        - file: out.log
          contains_matchin: "pattern"
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := unmarshalPlan(t, tc.input)
			if err == nil {
				t.Fatalf("expected error for unknown field %q, got nil", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error should mention %q, got: %v", tc.field, err)
			}
		})
	}
}

func TestSkipFieldsAccepted(t *testing.T) {
	input := `
specs:
  always-skip:
    skip: true
    command:
      cli: mytool
  skip-in-ci:
    skip_if:
      env_set: CI
    command:
      cli: mytool
  skip-on-windows:
    skip_if:
      env_value:
        GOOS: windows
    command:
      cli: mytool
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error parsing skip fields: %v", err)
	}
	if !plan.Specs["always-skip"].Skip {
		t.Error("expected skip=true for 'always-skip'")
	}
	if plan.Specs["skip-in-ci"].SkipIf.EnvSet != "CI" {
		t.Errorf("env_set = %q, want %q", plan.Specs["skip-in-ci"].SkipIf.EnvSet, "CI")
	}
	if plan.Specs["skip-on-windows"].SkipIf.EnvValue["GOOS"] != "windows" {
		t.Errorf("env_value[GOOS] = %q, want %q", plan.Specs["skip-on-windows"].SkipIf.EnvValue["GOOS"], "windows")
	}
}

func TestSkipTypoRejected(t *testing.T) {
	cases := []struct {
		name  string
		field string
		input string
	}{
		{
			name:  "skipp typo",
			field: "skipp",
			input: `
specs:
  test:
    skipp: true
    command:
      cli: mytool
`,
		},
		{
			name:  "env_sett typo",
			field: "env_sett",
			input: `
specs:
  test:
    skip_if:
      env_sett: CI
    command:
      cli: mytool
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := unmarshalPlan(t, tc.input)
			if err == nil {
				t.Fatalf("expected error for unknown field %q, got nil", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error should mention %q, got: %v", tc.field, err)
			}
		})
	}
}

func TestPlanSetupTeardownFieldsAccepted(t *testing.T) {
	input := `
setup:
  - cli: mkdir -p ./workdir
  - cli: echo ready
teardown:
  - cli: rm -rf ./workdir
specs:
  test:
    command:
      cli: mytool
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Setup) != 2 {
		t.Errorf("expected 2 setup commands, got %d", len(plan.Setup))
	}
	if plan.Setup[0].Cli != "mkdir -p ./workdir" {
		t.Errorf("setup[0].cli = %q", plan.Setup[0].Cli)
	}
	if len(plan.Teardown) != 1 {
		t.Errorf("expected 1 teardown command, got %d", len(plan.Teardown))
	}
}

func TestSpecSetupTeardownFieldsAccepted(t *testing.T) {
	input := `
specs:
  test:
    setup:
      - cli: echo seed > input.txt
    teardown:
      - cli: rm -f input.txt
    command:
      cli: mytool --input input.txt
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	spec := plan.Specs["test"]
	if len(spec.Setup) != 1 || spec.Setup[0].Cli != "echo seed > input.txt" {
		t.Errorf("spec setup not parsed correctly: %+v", spec.Setup)
	}
	if len(spec.Teardown) != 1 || spec.Teardown[0].Cli != "rm -f input.txt" {
		t.Errorf("spec teardown not parsed correctly: %+v", spec.Teardown)
	}
}

func TestSetupTeardownTyposRejected(t *testing.T) {
	typos := []struct {
		name  string
		field string
		input string
	}{
		{
			name:  "setupp at plan level",
			field: "setupp",
			input: "setupp:\n  - cli: x\nspecs:\n  t:\n    command:\n      cli: x\n",
		},
		{
			name:  "teardonw at spec level",
			field: "teardonw",
			input: "specs:\n  t:\n    teardonw:\n      - cli: x\n    command:\n      cli: x\n",
		},
		{
			name:  "unknown field inside setup command",
			field: "clli",
			input: "setup:\n  - clli: x\nspecs:\n  t:\n    command:\n      cli: x\n",
		},
	}
	for _, tc := range typos {
		t.Run(tc.name, func(t *testing.T) {
			_, err := unmarshalPlan(t, tc.input)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error should mention %q, got: %v", tc.field, err)
			}
		})
	}
}
