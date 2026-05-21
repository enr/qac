package qac

import (
	"strings"
	"testing"
)

// --- cli vs exe+args mutual exclusion ---

func TestCommand_CliAndExe_MutuallyExclusive_Spec(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	spec := Spec{
		Command: Command{
			Cli: "echo hello",
			Exe: "echo",
		},
	}
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if len(res.AllErrors()) == 0 {
		t.Error("expected config error when both cli and exe are set on a spec command")
	}
	found := false
	for _, b := range res.Blocks() {
		for _, entry := range b.Entries() {
			for _, err := range entry.Errors() {
				if strings.Contains(err.Error(), "cli") && strings.Contains(err.Error(), "exe") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("error message should mention both 'cli' and 'exe'")
	}
}

func TestCommand_CliAndExe_MutuallyExclusive_Setup(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	plan := TestPlan{
		Setup: []Command{
			{Cli: "echo setup", Exe: "echo"},
		},
		Specs: map[string]Spec{
			"s": {Command: Command{Cli: "echo ok"}},
		},
	}
	res := sut.Execute(plan)
	if len(res.AllErrors()) == 0 {
		t.Error("expected config error when both cli and exe are set on a setup command")
	}
}

// --- YAML: cli accepted, exe+args accepted, args typo rejected ---

func TestCommand_CliFieldAccepted(t *testing.T) {
	input := `
specs:
  test:
    command:
      cli: my-tool --flag value
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Specs["test"].Command.Cli != "my-tool --flag value" {
		t.Errorf("cli = %q, want %q", plan.Specs["test"].Command.Cli, "my-tool --flag value")
	}
}

func TestCommand_ExeAndArgsAccepted(t *testing.T) {
	input := `
specs:
  test:
    command:
      exe: my-tool
      args: [--flag, value]
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := plan.Specs["test"].Command
	if cmd.Exe != "my-tool" {
		t.Errorf("exe = %q, want %q", cmd.Exe, "my-tool")
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "--flag" || cmd.Args[1] != "value" {
		t.Errorf("args = %v, want [--flag value]", cmd.Args)
	}
}

func TestCommand_ArgTypoRejected(t *testing.T) {
	input := `
specs:
  test:
    command:
      exe: my-tool
      argss: [--flag]
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'argss', got nil")
	}
	if !strings.Contains(err.Error(), "argss") {
		t.Errorf("error should mention 'argss', got: %v", err)
	}
}
