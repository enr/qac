package qac

import (
	"testing"
)

// --- Cmd / ShellCmd ---

func TestCmd_SetsExeAndArgs(t *testing.T) {
	c := Cmd("echo", "hello", "world")
	if c.Exe != "echo" {
		t.Errorf("Exe = %q, want %q", c.Exe, "echo")
	}
	if len(c.Args) != 2 || c.Args[0] != "hello" || c.Args[1] != "world" {
		t.Errorf("Args = %v, want [hello world]", c.Args)
	}
	if c.Cli != "" {
		t.Errorf("Cmd() must not set Cli, got %q", c.Cli)
	}
}

func TestShellCmd_SetsCli(t *testing.T) {
	c := ShellCmd("echo hello | cat")
	if c.Cli != "echo hello | cat" {
		t.Errorf("Cli = %q, want %q", c.Cli, "echo hello | cat")
	}
	if c.Exe != "" {
		t.Errorf("ShellCmd() must not set Exe, got %q", c.Exe)
	}
}

// --- OutputMatcher constructors ---

func TestContains_SetsContainsAll(t *testing.T) {
	var a OutputAssertion
	Contains("foo", "bar")(&a)
	if len(a.ContainsAll) != 2 || a.ContainsAll[0] != "foo" || a.ContainsAll[1] != "bar" {
		t.Errorf("ContainsAll = %v, want [foo bar]", a.ContainsAll)
	}
}

func TestContainsAnyOf_SetsContainsAny(t *testing.T) {
	var a OutputAssertion
	ContainsAnyOf("foo", "bar")(&a)
	if len(a.ContainsAny) != 2 {
		t.Errorf("ContainsAny = %v, want [foo bar]", a.ContainsAny)
	}
}

func TestNotContains_SetsContainsNone(t *testing.T) {
	var a OutputAssertion
	NotContains("bad")(&a)
	if len(a.ContainsNone) != 1 || a.ContainsNone[0] != "bad" {
		t.Errorf("ContainsNone = %v, want [bad]", a.ContainsNone)
	}
}

func TestEquals_SetsEqualsTo(t *testing.T) {
	var a OutputAssertion
	Equals("exact")(&a)
	if a.EqualsTo != "exact" {
		t.Errorf("EqualsTo = %q, want %q", a.EqualsTo, "exact")
	}
}

func TestEmpty_SetsIsEmpty(t *testing.T) {
	var a OutputAssertion
	Empty()(&a)
	if a.IsEmpty == nil || !*a.IsEmpty {
		t.Errorf("IsEmpty should be true after Empty()")
	}
}

func TestStartsWith_SetsStartsWith(t *testing.T) {
	var a OutputAssertion
	StartsWith("pre")(&a)
	if a.StartsWith != "pre" {
		t.Errorf("StartsWith = %q, want %q", a.StartsWith, "pre")
	}
}

func TestEndsWith_SetsEndsWith(t *testing.T) {
	var a OutputAssertion
	EndsWith("suf")(&a)
	if a.EndsWith != "suf" {
		t.Errorf("EndsWith = %q, want %q", a.EndsWith, "suf")
	}
}

func TestMatches_SetsMatches(t *testing.T) {
	var a OutputAssertion
	Matches(`^\d+$`)(&a)
	if a.Matches != `^\d+$` {
		t.Errorf("Matches = %q, want %q", a.Matches, `^\d+$`)
	}
}

func TestNotMatches_SetsNotMatches(t *testing.T) {
	var a OutputAssertion
	NotMatches(`^\d+$`)(&a)
	if a.NotMatches != `^\d+$` {
		t.Errorf("NotMatches = %q, want %q", a.NotMatches, `^\d+$`)
	}
}

func TestHasLineCount_SetsLineCount(t *testing.T) {
	var a OutputAssertion
	HasLineCount(5)(&a)
	if a.LineCount == nil || *a.LineCount != 5 {
		t.Errorf("LineCount should be 5")
	}
}

func TestHasLineCountAtLeast_SetsLineCountGte(t *testing.T) {
	var a OutputAssertion
	HasLineCountAtLeast(3)(&a)
	if a.LineCountGte == nil || *a.LineCountGte != 3 {
		t.Errorf("LineCountGte should be 3")
	}
}

func TestContainsExactLine_SetsContainsLine(t *testing.T) {
	var a OutputAssertion
	ContainsExactLine("exact line")(&a)
	if a.ContainsLine != "exact line" {
		t.Errorf("ContainsLine = %q, want %q", a.ContainsLine, "exact line")
	}
}

// --- SpecBuilder ---

func TestSpecBuilder_Command(t *testing.T) {
	s := NewSpec().Command(Cmd("echo", "hi")).Build()
	if s.Command.Exe != "echo" || len(s.Command.Args) != 1 || s.Command.Args[0] != "hi" {
		t.Errorf("unexpected command: %+v", s.Command)
	}
}

func TestSpecBuilder_Description(t *testing.T) {
	s := NewSpec().Description("my spec").Build()
	if s.Description != "my spec" {
		t.Errorf("Description = %q, want %q", s.Description, "my spec")
	}
}

func TestSpecBuilder_Tags(t *testing.T) {
	s := NewSpec().Tags("fast", "smoke").Build()
	if len(s.Tags) != 2 || s.Tags[0] != "fast" || s.Tags[1] != "smoke" {
		t.Errorf("Tags = %v, want [fast smoke]", s.Tags)
	}
}

func TestSpecBuilder_Skip(t *testing.T) {
	s := NewSpec().Skip().Build()
	if !s.Skip {
		t.Error("Skip() should set Skip=true")
	}
}

func TestSpecBuilder_Retries(t *testing.T) {
	s := NewSpec().Retries(3).Build()
	if s.Retries != 3 {
		t.Errorf("Retries = %d, want 3", s.Retries)
	}
}

func TestSpecBuilder_RetryDelay(t *testing.T) {
	s := NewSpec().RetryDelay("500ms").Build()
	if s.RetryDelay != "500ms" {
		t.Errorf("RetryDelay = %q, want 500ms", s.RetryDelay)
	}
}

func TestSpecBuilder_ExpectStatus(t *testing.T) {
	s := NewSpec().ExpectStatus(0).Build()
	if s.Expectations.StatusAssertion.EqualsTo == nil || *s.Expectations.StatusAssertion.EqualsTo != 0 {
		t.Error("ExpectStatus(0) should set EqualsTo=0")
	}
}

func TestSpecBuilder_ExpectStatusGT(t *testing.T) {
	s := NewSpec().ExpectStatusGT(1).Build()
	if s.Expectations.StatusAssertion.GreaterThan == nil || *s.Expectations.StatusAssertion.GreaterThan != 1 {
		t.Error("ExpectStatusGT(1) should set GreaterThan=1")
	}
}

func TestSpecBuilder_ExpectStatusLT(t *testing.T) {
	s := NewSpec().ExpectStatusLT(10).Build()
	if s.Expectations.StatusAssertion.LessThan == nil || *s.Expectations.StatusAssertion.LessThan != 10 {
		t.Error("ExpectStatusLT(10) should set LessThan=10")
	}
}

func TestSpecBuilder_ExpectStdout(t *testing.T) {
	s := NewSpec().ExpectStdout(Contains("hello")).Build()
	a := s.Expectations.OutputAssertions.Stdout
	if len(a.ContainsAll) != 1 || a.ContainsAll[0] != "hello" {
		t.Errorf("Stdout.ContainsAll = %v, want [hello]", a.ContainsAll)
	}
}

func TestSpecBuilder_ExpectStderr(t *testing.T) {
	s := NewSpec().ExpectStderr(Contains("warn")).Build()
	a := s.Expectations.OutputAssertions.Stderr
	if len(a.ContainsAll) != 1 || a.ContainsAll[0] != "warn" {
		t.Errorf("Stderr.ContainsAll = %v, want [warn]", a.ContainsAll)
	}
}

func TestSpecBuilder_SetupTeardown(t *testing.T) {
	s := NewSpec().
		Setup(ShellCmd("touch /tmp/x")).
		Teardown(ShellCmd("rm /tmp/x")).
		Build()
	if len(s.Setup) != 1 || s.Setup[0].Cli != "touch /tmp/x" {
		t.Errorf("Setup = %v", s.Setup)
	}
	if len(s.Teardown) != 1 || s.Teardown[0].Cli != "rm /tmp/x" {
		t.Errorf("Teardown = %v", s.Teardown)
	}
}

// --- PlanBuilder ---

func TestPlanBuilder_SpecAdded(t *testing.T) {
	plan := NewPlan().
		Spec("echo", NewSpec().Command(Cmd("echo", "hi")).ExpectStatus(0)).
		Build()
	s, ok := plan.Specs["echo"]
	if !ok {
		t.Fatal("spec 'echo' not found in plan")
	}
	if s.Command.Exe != "echo" {
		t.Errorf("Exe = %q, want echo", s.Command.Exe)
	}
}

func TestPlanBuilder_SpecOrder(t *testing.T) {
	plan := NewPlan().
		Spec("first", NewSpec().Command(Cmd("echo", "1"))).
		Spec("second", NewSpec().Command(Cmd("echo", "2"))).
		Spec("third", NewSpec().Command(Cmd("echo", "3"))).
		Build()
	want := []string{"first", "second", "third"}
	if len(plan.specOrder) != 3 {
		t.Fatalf("specOrder = %v, want %v", plan.specOrder, want)
	}
	for i, name := range want {
		if plan.specOrder[i] != name {
			t.Errorf("specOrder[%d] = %q, want %q", i, plan.specOrder[i], name)
		}
	}
}

func TestPlanBuilder_Var(t *testing.T) {
	plan := NewPlan().Var("tool", "./bin/mytool").Var("env", "staging").Build()
	if plan.Vars["tool"] != "./bin/mytool" {
		t.Errorf("Vars[tool] = %q, want ./bin/mytool", plan.Vars["tool"])
	}
	if plan.Vars["env"] != "staging" {
		t.Errorf("Vars[env] = %q, want staging", plan.Vars["env"])
	}
}

func TestPlanBuilder_SetupTeardown(t *testing.T) {
	plan := NewPlan().
		Setup(ShellCmd("echo setup")).
		Teardown(ShellCmd("echo teardown")).
		Build()
	if len(plan.Setup) != 1 || plan.Setup[0].Cli != "echo setup" {
		t.Errorf("Setup = %v", plan.Setup)
	}
	if len(plan.Teardown) != 1 || plan.Teardown[0].Cli != "echo teardown" {
		t.Errorf("Teardown = %v", plan.Teardown)
	}
}

func TestPlanBuilder_AcceptsSpecValue(t *testing.T) {
	spec := NewSpec().Command(Cmd("echo", "direct")).Build()
	plan := NewPlan().Spec("direct", spec).Build()
	if _, ok := plan.Specs["direct"]; !ok {
		t.Error("spec 'direct' should be present when passed as Spec value")
	}
}

// --- Integration: builder + Execute ---

func TestBuilderIntegration_EchoSpec(t *testing.T) {
	plan := NewPlan().
		Spec("echo", NewSpec().
			Command(Cmd("echo", "hello")).
			ExpectStatus(0).
			ExpectStdout(Contains("hello"))).
		Build()
	report := NewLauncher().Execute(plan)
	if !report.Success() {
		t.Errorf("expected success, failed specs: %v", report.FailedSpecs())
	}
}

func TestBuilderIntegration_ShellCmd(t *testing.T) {
	plan := NewPlan().
		Spec("shell", NewSpec().
			Command(ShellCmd("echo shell-ok")).
			ExpectStatus(0).
			ExpectStdout(Contains("shell-ok"))).
		Build()
	report := NewLauncher().Execute(plan)
	if !report.Success() {
		t.Errorf("expected success, failed specs: %v", report.FailedSpecs())
	}
}

func TestBuilderIntegration_FailingSpec(t *testing.T) {
	plan := NewPlan().
		Spec("fail", NewSpec().
			Command(Cmd("echo", "hi")).
			ExpectStatus(99)).
		Build()
	report := NewLauncher().Execute(plan)
	if report.Success() {
		t.Error("expected failure for wrong expected exit code")
	}
	if len(report.FailedSpecs()) == 0 {
		t.Error("FailedSpecs should be non-empty")
	}
}

func TestBuilderIntegration_MultipleMatchers(t *testing.T) {
	plan := NewPlan().
		Spec("multi", NewSpec().
			Command(ShellCmd("printf 'line1\nline2\nline3\n'")).
			ExpectStatus(0).
			ExpectStdout(
				Contains("line1"),
				Contains("line3"),
				HasLineCountAtLeast(3),
			)).
		Build()
	report := NewLauncher().Execute(plan)
	if !report.Success() {
		t.Errorf("expected success, errors: %v", report.AllErrors())
	}
}
