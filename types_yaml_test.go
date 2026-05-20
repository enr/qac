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
        equals_to: "0"
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
