package qac

import (
	"strings"
	"testing"
)

// helpers shared by dryrun tests

func dryRunPlan(specs map[string]Spec, order []string) *TestExecutionReport {
	plan := TestPlan{Specs: specs, specOrder: order}
	return newLauncher(nil).DryRun(plan)
}

// --- DryRun: basic execution ---

func TestDryRun_ValidSpec_IsInfoEntry(t *testing.T) {
	report := dryRunPlan(
		map[string]Spec{"s": {id: "s", Command: Command{Cli: "echo hello"}}},
		[]string{"s"},
	)
	if len(report.AllErrors()) != 0 {
		t.Fatalf("expected no errors, got: %v", report.AllErrors())
	}
	found := false
	for _, b := range report.Blocks() {
		for _, e := range b.Entries() {
			if e.Kind() == InfoType && strings.Contains(e.Description(), "would execute") {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected an info entry containing 'would execute'")
	}
}

func TestDryRun_CommandSummary_Cli(t *testing.T) {
	report := dryRunPlan(
		map[string]Spec{"s": {id: "s", Command: Command{Cli: "my-tool --flag"}}},
		[]string{"s"},
	)
	for _, b := range report.Blocks() {
		for _, e := range b.Entries() {
			if e.Kind() == InfoType && strings.Contains(e.Description(), "my-tool --flag") {
				return
			}
		}
	}
	t.Error("expected 'my-tool --flag' in info description")
}

func TestDryRun_CommandSummary_ExeAndArgs(t *testing.T) {
	report := dryRunPlan(
		map[string]Spec{"s": {id: "s", Command: Command{Exe: "mytool", Args: []string{"--a", "b"}}}},
		[]string{"s"},
	)
	for _, b := range report.Blocks() {
		for _, e := range b.Entries() {
			if e.Kind() == InfoType && strings.Contains(e.Description(), "mytool") {
				return
			}
		}
	}
	t.Error("expected 'mytool' in info description")
}

func TestDryRun_NoCommandsExecuted(t *testing.T) {
	// The executor must never be called during a DryRun.
	called := false
	exec := &callTrackingExecutor{onExecute: func() { called = true }}
	plan := TestPlan{
		Specs:     map[string]Spec{"s": {id: "s", Command: Command{Cli: "echo hi"}}},
		specOrder: []string{"s"},
	}
	newLauncher(exec).DryRun(plan)
	if called {
		t.Error("DryRun must not call the executor")
	}
}

// --- DryRun: config error detection ---

func TestDryRun_CliAndExe_IsError(t *testing.T) {
	report := dryRunPlan(
		map[string]Spec{"s": {id: "s", Command: Command{Cli: "foo", Exe: "bar"}}},
		[]string{"s"},
	)
	if len(report.AllErrors()) == 0 {
		t.Error("expected config error when both cli and exe are set")
	}
}

func TestDryRun_StdinAndStdinFile_IsError(t *testing.T) {
	report := dryRunPlan(
		map[string]Spec{"s": {id: "s", Command: Command{Cli: "foo", Stdin: "x", StdinFile: "f.txt"}}},
		[]string{"s"},
	)
	if len(report.AllErrors()) == 0 {
		t.Error("expected config error when both stdin and stdin_file are set")
	}
}

func TestDryRun_InvalidRetryDelay_IsError(t *testing.T) {
	report := dryRunPlan(
		map[string]Spec{"s": {id: "s", Command: Command{Cli: "echo"}, RetryDelay: "not-a-duration"}},
		[]string{"s"},
	)
	if len(report.AllErrors()) == 0 {
		t.Error("expected config error for invalid retry_delay")
	}
}

func TestDryRun_SetupCommandError_Reported(t *testing.T) {
	spec := Spec{
		id:      "s",
		Command: Command{Cli: "echo"},
		Setup:   []Command{{Cli: "foo", Exe: "bar"}},
	}
	report := dryRunPlan(map[string]Spec{"s": spec}, []string{"s"})
	if len(report.AllErrors()) == 0 {
		t.Error("expected config error for invalid setup command")
	}
}

func TestDryRun_PlanSetupCommandError_Reported(t *testing.T) {
	plan := TestPlan{
		Setup:     []Command{{Cli: "foo", Exe: "bar"}},
		Specs:     map[string]Spec{"s": {id: "s", Command: Command{Cli: "echo"}}},
		specOrder: []string{"s"},
	}
	report := newLauncher(nil).DryRun(plan)
	if len(report.AllErrors()) == 0 {
		t.Error("expected config error for invalid plan-level setup command")
	}
}

// --- DryRun: skip propagation ---

func TestDryRun_StaticSkip_IsSkipped(t *testing.T) {
	report := dryRunPlan(
		map[string]Spec{"s": {id: "s", Skip: true, Command: Command{Cli: "echo"}}},
		[]string{"s"},
	)
	for _, b := range report.Blocks() {
		if b.Phase() == "s" && b.Skipped() {
			return
		}
	}
	t.Error("expected skipped block for spec with skip:true")
}

func TestDryRun_TagFilter_SkipsNonMatchingSpec(t *testing.T) {
	plan := TestPlan{
		Specs: map[string]Spec{
			"fast-spec": {id: "fast-spec", Tags: []string{"fast"}, Command: Command{Cli: "echo"}},
			"slow-spec": {id: "slow-spec", Tags: []string{"slow"}, Command: Command{Cli: "echo"}},
		},
		specOrder: []string{"fast-spec", "slow-spec"},
	}
	report := newLauncher(nil).DryRun(plan, WithTags("fast"))
	for _, b := range report.Blocks() {
		if b.Phase() == "slow-spec" && !b.Skipped() {
			t.Error("slow-spec should be skipped when WithTags('fast') is set")
		}
	}
}

// --- DryRun: block structure ---

func TestDryRun_BlockIndexAndTotal(t *testing.T) {
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {id: "a", Command: Command{Cli: "echo"}},
			"b": {id: "b", Command: Command{Cli: "echo"}},
		},
		specOrder: []string{"a", "b"},
	}
	report := newLauncher(nil).DryRun(plan)
	indexed := []*ReportBlock{}
	for _, b := range report.Blocks() {
		if b.Index() > 0 {
			indexed = append(indexed, b)
		}
	}
	if len(indexed) != 2 {
		t.Fatalf("expected 2 indexed blocks, got %d", len(indexed))
	}
	if indexed[0].Total() != 2 || indexed[1].Total() != 2 {
		t.Error("expected Total()=2 for both blocks")
	}
}

// --- ListSpecs ---

func TestListSpecs_ReturnsAllInOrder(t *testing.T) {
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {id: "a", Command: Command{Cli: "echo"}},
			"b": {id: "b", Command: Command{Cli: "echo"}},
			"c": {id: "c", Command: Command{Cli: "echo"}},
		},
		specOrder: []string{"a", "b", "c"},
	}
	got := newLauncher(nil).ListSpecs(plan)
	want := []string{"a", "b", "c"}
	if !sliceEqual(got, want) {
		t.Errorf("ListSpecs = %v, want %v", got, want)
	}
}

func TestListSpecs_ExcludesStaticSkip(t *testing.T) {
	plan := TestPlan{
		Specs: map[string]Spec{
			"run":  {id: "run", Command: Command{Cli: "echo"}},
			"skip": {id: "skip", Skip: true, Command: Command{Cli: "echo"}},
		},
		specOrder: []string{"run", "skip"},
	}
	got := newLauncher(nil).ListSpecs(plan)
	if len(got) != 1 || got[0] != "run" {
		t.Errorf("ListSpecs = %v, want [run]", got)
	}
}

func TestListSpecs_TagFilter(t *testing.T) {
	plan := TestPlan{
		Specs: map[string]Spec{
			"fast": {id: "fast", Tags: []string{"fast"}, Command: Command{Cli: "echo"}},
			"slow": {id: "slow", Tags: []string{"slow"}, Command: Command{Cli: "echo"}},
		},
		specOrder: []string{"fast", "slow"},
	}
	got := newLauncher(nil).ListSpecs(plan, WithTags("fast"))
	if len(got) != 1 || got[0] != "fast" {
		t.Errorf("ListSpecs with WithTags('fast') = %v, want [fast]", got)
	}
}

func TestListSpecs_EmptyPlan(t *testing.T) {
	got := newLauncher(nil).ListSpecs(TestPlan{})
	if len(got) != 0 {
		t.Errorf("ListSpecs on empty plan = %v, want []", got)
	}
}

// --- helpers ---

type callTrackingExecutor struct {
	onExecute func()
}

func (e *callTrackingExecutor) execute(_ Command) executionResult {
	if e.onExecute != nil {
		e.onExecute()
	}
	return executionResult{success: true}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
