package qac

import (
	"strings"
	"testing"
)

// --- tagSkipReason unit tests ---

func TestTagSkipReason_NoFilter_NoTags(t *testing.T) {
	if r := tagSkipReason(Spec{}, runConfig{}); r != "" {
		t.Errorf("expected no skip reason, got %q", r)
	}
}

func TestTagSkipReason_NoFilter_SpecHasTags(t *testing.T) {
	spec := Spec{Tags: []string{"fast"}}
	if r := tagSkipReason(spec, runConfig{}); r != "" {
		t.Errorf("expected no skip reason when no filter set, got %q", r)
	}
}

func TestTagSkipReason_WithTags_Match(t *testing.T) {
	spec := Spec{Tags: []string{"fast", "unit"}}
	if r := tagSkipReason(spec, runConfig{withTags: []string{"fast"}}); r != "" {
		t.Errorf("expected no skip reason when tag matches, got %q", r)
	}
}

func TestTagSkipReason_WithTags_NoMatch(t *testing.T) {
	spec := Spec{Tags: []string{"slow"}}
	if r := tagSkipReason(spec, runConfig{withTags: []string{"fast"}}); r == "" {
		t.Error("expected skip reason when spec has no matching tag")
	}
}

func TestTagSkipReason_WithTags_EmptySpec(t *testing.T) {
	if r := tagSkipReason(Spec{}, runConfig{withTags: []string{"fast"}}); r == "" {
		t.Error("expected skip reason when spec has no tags and withTags filter is set")
	}
}

func TestTagSkipReason_WithTags_MultipleOptions(t *testing.T) {
	// WithTags("fast") AND WithTags("unit") behave as OR — spec needs at least one
	spec := Spec{Tags: []string{"unit"}}
	if r := tagSkipReason(spec, runConfig{withTags: []string{"fast", "unit"}}); r != "" {
		t.Errorf("expected no skip reason (OR semantics), got %q", r)
	}
}

func TestTagSkipReason_SkipTags_Match(t *testing.T) {
	spec := Spec{Tags: []string{"slow", "network"}}
	if r := tagSkipReason(spec, runConfig{skipTags: []string{"network"}}); r == "" {
		t.Error("expected skip reason when spec has a skipped tag")
	}
}

func TestTagSkipReason_SkipTags_NoMatch(t *testing.T) {
	spec := Spec{Tags: []string{"fast"}}
	if r := tagSkipReason(spec, runConfig{skipTags: []string{"network"}}); r != "" {
		t.Errorf("expected no skip reason when spec has no skipped tags, got %q", r)
	}
}

func TestTagSkipReason_SkipTags_TakesPrecedence(t *testing.T) {
	// spec has "fast" (matches withTags) AND "network" (matches skipTags)
	// skipTags should win
	spec := Spec{Tags: []string{"fast", "network"}}
	cfg := runConfig{withTags: []string{"fast"}, skipTags: []string{"network"}}
	if r := tagSkipReason(spec, cfg); r == "" {
		t.Error("expected skip reason: skipTags should take precedence over withTags")
	}
}

// --- YAML parsing ---

func TestTagsFieldAccepted(t *testing.T) {
	input := `
specs:
  fast-spec:
    tags: [fast, unit]
    command:
      cli: mytool
  slow-spec:
    tags:
      - slow
      - network
    command:
      cli: mytool
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := plan.Specs["fast-spec"].Tags; len(got) != 2 || got[0] != "fast" || got[1] != "unit" {
		t.Errorf("fast-spec tags = %v, want [fast unit]", got)
	}
	if got := plan.Specs["slow-spec"].Tags; len(got) != 2 || got[0] != "slow" || got[1] != "network" {
		t.Errorf("slow-spec tags = %v, want [slow network]", got)
	}
}

func TestTagsTypoRejected(t *testing.T) {
	input := `
specs:
  test:
    taggs: [fast]
    command:
      cli: mytool
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'taggs', got nil")
	}
	if !strings.Contains(err.Error(), "taggs") {
		t.Errorf("error should mention 'taggs', got: %v", err)
	}
}

// --- Launcher integration ---

func TestExecute_WithTags_OnlyMatchingRun(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	plan := TestPlan{
		Specs: map[string]Spec{
			"fast-spec": {
				Tags:         []string{"fast"},
				Command:      Command{Exe: "true"},
				Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}},
			},
			"slow-spec": {
				Tags:         []string{"slow"},
				Command:      Command{Exe: "true"},
				Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}},
			},
		},
	}
	res := sut.Execute(plan, WithTags("fast"))
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
	skipped := 0
	for _, b := range res.Blocks() {
		if b.Skipped() {
			skipped++
		}
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped spec (slow-spec), got %d", skipped)
	}
}

func TestExecute_SkipTags_ExcludesMatching(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	plan := TestPlan{
		Specs: map[string]Spec{
			"fast-spec": {
				Tags:         []string{"fast"},
				Command:      Command{Exe: "true"},
				Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}},
			},
			"network-spec": {
				Tags:         []string{"network"},
				Command:      Command{Exe: "true"},
				Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}},
			},
		},
	}
	res := sut.Execute(plan, SkipTags("network"))
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
	skipped := 0
	for _, b := range res.Blocks() {
		if b.Skipped() {
			skipped++
		}
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped spec (network-spec), got %d", skipped)
	}
}

func TestExecute_NoOptions_AllSpecsRun(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {Tags: []string{"fast"}, Command: Command{Exe: "true"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}}},
			"b": {Tags: []string{"slow"}, Command: Command{Exe: "true"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}}},
			"c": {Command: Command{Exe: "true"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}}},
		},
	}
	res := sut.Execute(plan)
	for _, b := range res.Blocks() {
		if b.Skipped() {
			t.Errorf("spec %q should not be skipped when no options passed", b.Phase())
		}
	}
}

func TestExecute_WithAndSkipTags_Combined(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	plan := TestPlan{
		Specs: map[string]Spec{
			"fast-local": {
				Tags:         []string{"fast"},
				Command:      Command{Exe: "true"},
				Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}},
			},
			"fast-network": {
				Tags:         []string{"fast", "network"},
				Command:      Command{Exe: "true"},
				Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}},
			},
			"slow-local": {
				Tags:         []string{"slow"},
				Command:      Command{Exe: "true"},
				Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}},
			},
		},
	}
	// want: fast specs, but not those tagged network → only fast-local runs
	res := sut.Execute(plan, WithTags("fast"), SkipTags("network"))
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
	skipped := 0
	for _, b := range res.Blocks() {
		if b.Skipped() {
			skipped++
		}
	}
	if skipped != 2 {
		t.Errorf("expected 2 skipped specs (fast-network + slow-local), got %d", skipped)
	}
}

func TestExecute_UntaggedSpecSkippedByWithTags(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	plan := TestPlan{
		Specs: map[string]Spec{
			"tagged":   {Tags: []string{"fast"}, Command: Command{Exe: "true"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}}},
			"untagged": {Command: Command{Exe: "true"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: &zero}}},
		},
	}
	res := sut.Execute(plan, WithTags("fast"))
	skipped := 0
	for _, b := range res.Blocks() {
		if b.Skipped() {
			skipped++
		}
	}
	if skipped != 1 {
		t.Errorf("expected untagged spec to be skipped by WithTags filter, got %d skipped", skipped)
	}
}
