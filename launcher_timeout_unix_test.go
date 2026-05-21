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
