package qac

import (
	"testing"
	"time"
)

// A minimal plan with two specs to verify index/total/duration tracking.
var twoSpecPlan = TestPlan{
	Specs: map[string]Spec{
		"alpha": {id: "alpha", Command: Command{Cli: "true"}},
		"beta":  {id: "beta", Command: Command{Cli: "true"}},
	},
	specOrder: []string{"alpha", "beta"},
}

func TestBlockIndexAndTotal(t *testing.T) {
	launcher := newLauncher(&fixedValueExecutor{exitCode: 0})
	report := launcher.Execute(twoSpecPlan)

	indexed := []*ReportBlock{}
	for _, b := range report.Blocks() {
		if b.Index() > 0 {
			indexed = append(indexed, b)
		}
	}
	if len(indexed) != 2 {
		t.Fatalf("expected 2 indexed blocks, got %d", len(indexed))
	}
	if indexed[0].Index() != 1 || indexed[0].Total() != 2 {
		t.Errorf("first block: want index=1 total=2, got index=%d total=%d", indexed[0].Index(), indexed[0].Total())
	}
	if indexed[1].Index() != 2 || indexed[1].Total() != 2 {
		t.Errorf("second block: want index=2 total=2, got index=%d total=%d", indexed[1].Index(), indexed[1].Total())
	}
}

func TestBlockDurationIsSet(t *testing.T) {
	launcher := newLauncher(&fixedValueExecutor{exitCode: 0})
	report := launcher.Execute(twoSpecPlan)

	for _, b := range report.Blocks() {
		if b.Index() > 0 && b.Duration() <= 0 {
			t.Errorf("block %q: expected positive duration, got %v", b.Phase(), b.Duration())
		}
	}
}

func TestBlockStartedAtIsSet(t *testing.T) {
	before := time.Now()
	launcher := newLauncher(&fixedValueExecutor{exitCode: 0})
	report := launcher.Execute(twoSpecPlan)
	after := time.Now()

	for _, b := range report.Blocks() {
		if b.Index() > 0 {
			if b.StartedAt().Before(before) || b.StartedAt().After(after) {
				t.Errorf("block %q: startedAt %v outside test window [%v, %v]", b.Phase(), b.StartedAt(), before, after)
			}
		}
	}
}

func TestNonSpecBlockHasZeroIndex(t *testing.T) {
	// plan-level preconditions block should have index=0
	plan := TestPlan{
		Preconditions: Preconditions{
			FileSystemAssertions: []FileSystemAssertion{
				{File: "nonexistent_xyz_abc.txt", Exists: boolPtr(false)},
			},
		},
		Specs: map[string]Spec{
			"only": {id: "only", Command: Command{Cli: "true"}},
		},
		specOrder: []string{"only"},
	}
	launcher := newLauncher(&fixedValueExecutor{exitCode: 0})
	report := launcher.Execute(plan)

	for _, b := range report.Blocks() {
		if b.Phase() == "preconditions" && b.Index() != 0 {
			t.Errorf("preconditions block should have index=0, got %d", b.Index())
		}
	}
}

func TestSkippedSpecBlockStatus(t *testing.T) {
	// A spec whose precondition requires a file that doesn't exist → SKIP
	plan := TestPlan{
		Specs: map[string]Spec{
			"needs-file": {
				id: "needs-file",
				Preconditions: Preconditions{
					FileSystemAssertions: []FileSystemAssertion{
						{File: "nonexistent_xyz_abc.txt", Exists: boolPtr(true)},
					},
				},
				Command: Command{Cli: "true"},
			},
		},
		specOrder: []string{"needs-file"},
	}
	launcher := newLauncher(&fixedValueExecutor{exitCode: 0})
	report := launcher.Execute(plan)

	for _, b := range report.Blocks() {
		if b.Phase() == "needs-file" {
			if blockStatus(b) != "SKIP" {
				t.Errorf("expected SKIP for block with failed precondition, got %q", blockStatus(b))
			}
			return
		}
	}
	t.Error("block 'needs-file' not found in report")
}
