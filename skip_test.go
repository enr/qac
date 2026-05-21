package qac

import (
	"os"
	"testing"
)

// --- specSkipReason unit tests ---

func TestSkipReason_StaticSkipTrue(t *testing.T) {
	if specSkipReason(Spec{Skip: true}) == "" {
		t.Error("expected non-empty reason for skip=true")
	}
}

func TestSkipReason_StaticSkipFalse(t *testing.T) {
	if got := specSkipReason(Spec{Skip: false}); got != "" {
		t.Errorf("expected empty reason for skip=false, got %q", got)
	}
}

func TestSkipReason_NoConditions(t *testing.T) {
	if got := specSkipReason(Spec{}); got != "" {
		t.Errorf("expected empty reason for spec with no skip fields, got %q", got)
	}
}

func TestSkipReason_EnvSet_Defined(t *testing.T) {
	t.Setenv("QAC_TEST_ENV_SET", "1")
	if specSkipReason(Spec{SkipIf: SkipCondition{EnvSet: "QAC_TEST_ENV_SET"}}) == "" {
		t.Error("expected non-empty reason when env_set variable is defined")
	}
}

func TestSkipReason_EnvSet_EmptyValue(t *testing.T) {
	// Defined with empty string still counts as "set".
	t.Setenv("QAC_TEST_ENV_SET_EMPTY", "")
	if specSkipReason(Spec{SkipIf: SkipCondition{EnvSet: "QAC_TEST_ENV_SET_EMPTY"}}) == "" {
		t.Error("expected non-empty reason when env_set variable is set to empty string")
	}
}

func TestSkipReason_EnvSet_NotDefined(t *testing.T) {
	os.Unsetenv("QAC_TEST_ENV_SET_ABSENT")
	if got := specSkipReason(Spec{SkipIf: SkipCondition{EnvSet: "QAC_TEST_ENV_SET_ABSENT"}}); got != "" {
		t.Errorf("expected empty reason when env_set variable is absent, got %q", got)
	}
}

func TestSkipReason_EnvValue_Matches(t *testing.T) {
	t.Setenv("QAC_TEST_GOOS", "windows")
	spec := Spec{SkipIf: SkipCondition{EnvValue: map[string]string{"QAC_TEST_GOOS": "windows"}}}
	if specSkipReason(spec) == "" {
		t.Error("expected non-empty reason when env_value matches")
	}
}

func TestSkipReason_EnvValue_NoMatch(t *testing.T) {
	t.Setenv("QAC_TEST_GOOS", "linux")
	spec := Spec{SkipIf: SkipCondition{EnvValue: map[string]string{"QAC_TEST_GOOS": "windows"}}}
	if got := specSkipReason(spec); got != "" {
		t.Errorf("expected empty reason when env_value does not match, got %q", got)
	}
}

func TestSkipReason_EnvValue_OneOfMultipleMatches(t *testing.T) {
	t.Setenv("QAC_TEST_GOOS", "linux")
	t.Setenv("QAC_TEST_GOARCH", "arm64")
	spec := Spec{SkipIf: SkipCondition{EnvValue: map[string]string{
		"QAC_TEST_GOOS":   "windows",
		"QAC_TEST_GOARCH": "arm64",
	}}}
	if specSkipReason(spec) == "" {
		t.Error("expected non-empty reason when at least one env_value entry matches")
	}
}

func TestSkipReason_EnvValue_NoneMatch(t *testing.T) {
	t.Setenv("QAC_TEST_GOOS", "linux")
	t.Setenv("QAC_TEST_GOARCH", "amd64")
	spec := Spec{SkipIf: SkipCondition{EnvValue: map[string]string{
		"QAC_TEST_GOOS":   "windows",
		"QAC_TEST_GOARCH": "arm64",
	}}}
	if got := specSkipReason(spec); got != "" {
		t.Errorf("expected empty reason when no env_value entry matches, got %q", got)
	}
}

func TestSkipReason_StaticSkipTakesPrecedence(t *testing.T) {
	// skip: true short-circuits before env checks.
	os.Unsetenv("QAC_TEST_ABSENT")
	spec := Spec{
		Skip:   true,
		SkipIf: SkipCondition{EnvSet: "QAC_TEST_ABSENT"},
	}
	reason := specSkipReason(spec)
	if reason == "" {
		t.Error("expected non-empty reason")
	}
	if reason != "skip: true" {
		t.Errorf("expected reason to be 'skip: true', got %q", reason)
	}
}

// --- Launcher integration: skipped specs appear as SkippedType in the report ---

func TestLauncher_StaticSkip_ReportedAsSkipped(t *testing.T) {
	e := &fixedValueExecutor{exitCode: 0}
	sut := newLauncher(e)

	specs := map[string]Spec{
		"skipped-spec": {Skip: true, Command: Command{Exe: "true"}},
	}
	res := sut.Execute(TestPlan{Specs: specs})

	if len(res.AllErrors()) != 0 {
		t.Errorf("expected 0 errors for a skipped spec, got: %v", res.AllErrors())
	}
	var found bool
	for _, b := range res.Blocks() {
		for _, entry := range b.Entries() {
			if entry.Kind() == SkippedType {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected at least one SkippedType entry in the report")
	}
}

func TestLauncher_SkipIfEnvSet_EnvPresent_Skipped(t *testing.T) {
	t.Setenv("QAC_TEST_CI", "true")
	e := &fixedValueExecutor{exitCode: 0}
	sut := newLauncher(e)

	specs := map[string]Spec{
		"net-test": {
			SkipIf:  SkipCondition{EnvSet: "QAC_TEST_CI"},
			Command: Command{Exe: "curl"},
		},
	}
	res := sut.Execute(TestPlan{Specs: specs})

	if len(res.AllErrors()) != 0 {
		t.Errorf("expected 0 errors, got: %v", res.AllErrors())
	}
	for _, b := range res.Blocks() {
		for _, entry := range b.Entries() {
			if entry.Kind() == SkippedType {
				return // found it
			}
		}
	}
	t.Error("expected a SkippedType entry when env_set condition is met")
}

func TestLauncher_SkipIfEnvSet_EnvAbsent_Runs(t *testing.T) {
	os.Unsetenv("QAC_TEST_CI_ABSENT")
	exitCode := 0
	e := &fixedValueExecutor{success: true, exitCode: exitCode}
	sut := newLauncher(e)

	specs := map[string]Spec{
		"net-test": {
			SkipIf: SkipCondition{EnvSet: "QAC_TEST_CI_ABSENT"},
			Command: Command{Exe: "true"},
			Expectations: Expectations{
				StatusAssertion: StatusAssertion{EqualsTo: &exitCode},
			},
		},
	}
	res := sut.Execute(TestPlan{Specs: specs})

	if len(res.AllErrors()) != 0 {
		t.Errorf("expected 0 errors when skip condition is not met, got: %v", res.AllErrors())
	}
	for _, b := range res.Blocks() {
		for _, entry := range b.Entries() {
			if entry.Kind() == SkippedType {
				t.Error("spec should have run, not been skipped")
			}
		}
	}
}

func TestLauncher_SkipIfEnvValue_Matches_Skipped(t *testing.T) {
	t.Setenv("QAC_TEST_PLATFORM", "windows")
	e := &fixedValueExecutor{}
	sut := newLauncher(e)

	specs := map[string]Spec{
		"windows-only": {
			SkipIf:  SkipCondition{EnvValue: map[string]string{"QAC_TEST_PLATFORM": "windows"}},
			Command: Command{Exe: "powershell"},
		},
	}
	res := sut.Execute(TestPlan{Specs: specs})

	for _, b := range res.Blocks() {
		for _, entry := range b.Entries() {
			if entry.Kind() == SkippedType {
				return
			}
		}
	}
	t.Error("expected a SkippedType entry when env_value condition is met")
}
