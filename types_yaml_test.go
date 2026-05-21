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
