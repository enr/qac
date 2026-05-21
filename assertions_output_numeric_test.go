package qac

import (
	"strings"
	"testing"
)

// --- line_count ---

func TestLineCount_ExactMatch(t *testing.T) {
	n := 3
	r := (&OutputAssertion{id: "stdout", LineCount: &n}).verify(outputCtx("a\nb\nc", ""))
	if !r.Success() {
		t.Errorf("expected success for 3-line output, got: %v", r.Errors())
	}
}

func TestLineCount_TooFew(t *testing.T) {
	n := 5
	r := (&OutputAssertion{id: "stdout", LineCount: &n}).verify(outputCtx("a\nb", ""))
	if r.Success() {
		t.Error("expected failure when line count is less than expected")
	}
}

func TestLineCount_TooMany(t *testing.T) {
	n := 1
	r := (&OutputAssertion{id: "stdout", LineCount: &n}).verify(outputCtx("a\nb\nc", ""))
	if r.Success() {
		t.Error("expected failure when line count exceeds expected")
	}
}

func TestLineCount_Zero_EmptyOutput(t *testing.T) {
	n := 0
	r := (&OutputAssertion{id: "stdout", LineCount: &n}).verify(outputCtx("", ""))
	if !r.Success() {
		t.Errorf("expected success for empty output with line_count 0, got: %v", r.Errors())
	}
}

func TestLineCount_Zero_NonEmptyOutput(t *testing.T) {
	n := 0
	r := (&OutputAssertion{id: "stdout", LineCount: &n}).verify(outputCtx("some output", ""))
	if r.Success() {
		t.Error("expected failure: non-empty output with line_count 0")
	}
}

func TestLineCount_TrailingNewlineIgnored(t *testing.T) {
	// TrimSpace strips trailing newline, so "a\nb\n" == "a\nb" == 2 lines
	n := 2
	r := (&OutputAssertion{id: "stdout", LineCount: &n}).verify(outputCtx("a\nb\n", ""))
	if !r.Success() {
		t.Errorf("expected success; trailing newline should not add a line, got: %v", r.Errors())
	}
}

// --- line_count_gte ---

func TestLineCountGte_Exact(t *testing.T) {
	n := 3
	r := (&OutputAssertion{id: "stdout", LineCountGte: &n}).verify(outputCtx("a\nb\nc", ""))
	if !r.Success() {
		t.Errorf("expected success when count == gte, got: %v", r.Errors())
	}
}

func TestLineCountGte_More(t *testing.T) {
	n := 2
	r := (&OutputAssertion{id: "stdout", LineCountGte: &n}).verify(outputCtx("a\nb\nc\nd", ""))
	if !r.Success() {
		t.Errorf("expected success when count > gte, got: %v", r.Errors())
	}
}

func TestLineCountGte_Fail(t *testing.T) {
	n := 5
	r := (&OutputAssertion{id: "stdout", LineCountGte: &n}).verify(outputCtx("a\nb", ""))
	if r.Success() {
		t.Error("expected failure when line count < gte")
	}
}

// --- contains_line ---

func TestContainsLine_Match(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsLine: "Status: OK"}).
		verify(outputCtx("Starting...\nStatus: OK\nDone.", ""))
	if !r.Success() {
		t.Errorf("expected success when line matches exactly, got: %v", r.Errors())
	}
}

func TestContainsLine_NoMatch(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsLine: "Status: OK"}).
		verify(outputCtx("Starting...\nStatus: FAIL\nDone.", ""))
	if r.Success() {
		t.Error("expected failure when no line equals contains_line")
	}
}

func TestContainsLine_PartialDoesNotMatch(t *testing.T) {
	// "Status: OK extra" should not satisfy contains_line: "Status: OK"
	r := (&OutputAssertion{id: "stdout", ContainsLine: "Status: OK"}).
		verify(outputCtx("Status: OK extra", ""))
	if r.Success() {
		t.Error("expected failure: partial line match should not count")
	}
}

func TestContainsLine_SingleLine(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsLine: "hello"}).verify(outputCtx("hello", ""))
	if !r.Success() {
		t.Errorf("expected success for single-line match, got: %v", r.Errors())
	}
}

func TestContainsLine_EmptyOutput(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", ContainsLine: "Status: OK"}).verify(outputCtx("", ""))
	if r.Success() {
		t.Error("expected failure when output is empty")
	}
}

// --- YAML parsing ---

func TestLineCountFieldsAccepted(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      output:
        stdout:
          line_count: 5
          line_count_gte: 3
          contains_line: "Status: OK"
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stdout := plan.Specs["test"].Expectations.OutputAssertions.Stdout
	if stdout.LineCount == nil || *stdout.LineCount != 5 {
		t.Errorf("line_count = %v, want 5", stdout.LineCount)
	}
	if stdout.LineCountGte == nil || *stdout.LineCountGte != 3 {
		t.Errorf("line_count_gte = %v, want 3", stdout.LineCountGte)
	}
	if stdout.ContainsLine != "Status: OK" {
		t.Errorf("contains_line = %q, want %q", stdout.ContainsLine, "Status: OK")
	}
}

func TestLineCountTyposRejected(t *testing.T) {
	cases := []struct {
		name  string
		field string
		input string
	}{
		{
			name:  "line_cont typo",
			field: "line_cont",
			input: `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      output:
        stdout:
          line_cont: 5
`,
		},
		{
			name:  "line_count_gt typo",
			field: "line_count_gt",
			input: `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      output:
        stdout:
          line_count_gt: 3
`,
		},
		{
			name:  "contains_lien typo",
			field: "contains_lien",
			input: `
specs:
  test:
    command:
      cli: my-tool
    expectations:
      output:
        stdout:
          contains_lien: "Status: OK"
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
