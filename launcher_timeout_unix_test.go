//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package qac

import (
	"testing"
)

func TestTimeoutKillsHangingCommand(t *testing.T) {
	sut := NewLauncher()

	spec := Spec{
		Command: Command{
			Cli:     "sleep 60",
			Timeout: "100ms",
		},
		Expectations: Expectations{},
	}

	specs := map[string]Spec{"slow": spec}
	plan := TestPlan{Specs: specs}

	res := sut.Execute(plan)

	var timedOutBlock *ReportBlock
	for _, b := range res.Blocks() {
		if b.TimedOut() {
			timedOutBlock = b
			break
		}
	}
	if timedOutBlock == nil {
		t.Fatal("expected a timed-out block but none found")
	}

	errs := res.AllErrors()
	if len(errs) == 0 {
		t.Fatal("expected at least one error for the timed-out command")
	}
}

func TestNoTimeoutWhenCommandFinishesInTime(t *testing.T) {
	sut := NewLauncher()

	zero := 0
	spec := Spec{
		Command: Command{
			Cli:     "true",
			Timeout: "5s",
		},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
		},
	}

	specs := map[string]Spec{"fast": spec}
	plan := TestPlan{Specs: specs}

	res := sut.Execute(plan)

	for _, b := range res.Blocks() {
		if b.TimedOut() {
			t.Errorf("block %q should not have timed out", b.Phase())
		}
	}
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestTimeoutCommandExitsBeforeTimeout_OutputPreserved(t *testing.T) {
	// A command that produces output and exits before the timeout must have its
	// stdout and exit code captured correctly — not truncated by timeout machinery.
	zero := 0
	spec := Spec{
		Command: Command{
			Cli:     "echo qac-hello",
			Timeout: "5s",
		},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
			OutputAssertions: OutputAssertions{
				Stdout: OutputAssertion{EqualsTo: "qac-hello"},
			},
		},
	}
	res := NewLauncher().Execute(TestPlan{Specs: map[string]Spec{"fast": spec}})
	for _, b := range res.Blocks() {
		if b.TimedOut() {
			t.Errorf("block %q should not have timed out", b.Phase())
		}
	}
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestLauncher_EnvOverridesParentEnvironment(t *testing.T) {
	// mergeEnv appends custom entries after os.Environ(); the last occurrence of
	// a duplicate key wins on Linux/macOS. Verify the child process sees the
	// custom value, not the one inherited from the parent.
	t.Setenv("QAC_TEST_ENV_OVERRIDE", "parent_value")
	zero := 0
	spec := Spec{
		Command: Command{
			Cli:     "echo $QAC_TEST_ENV_OVERRIDE",
			Timeout: "5s", // forces executeDirect, which calls mergeEnv
			Env:     map[string]string{"QAC_TEST_ENV_OVERRIDE": "custom_value"},
		},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
			OutputAssertions: OutputAssertions{
				Stdout: OutputAssertion{EqualsTo: "custom_value"},
			},
		},
	}
	res := NewLauncher().Execute(TestPlan{Specs: map[string]Spec{"env-override": spec}})
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
}
