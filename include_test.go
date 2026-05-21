package qac

import (
	"path/filepath"
	"strings"
	"testing"
)

// --- YAML parsing ---

func TestIncludeFieldAccepted(t *testing.T) {
	input := `
include:
  - ./common/base.yaml
specs:
  test:
    command:
      cli: mytool
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Include) != 1 || plan.Include[0] != "./common/base.yaml" {
		t.Errorf("include = %v, want [./common/base.yaml]", plan.Include)
	}
}

func TestIncludeTypoRejected(t *testing.T) {
	input := `
inclde:
  - ./base.yaml
specs:
  test:
    command:
      cli: mytool
`
	_, err := unmarshalPlan(t, input)
	if err == nil {
		t.Fatal("expected error for unknown field 'inclde', got nil")
	}
	if !strings.Contains(err.Error(), "inclde") {
		t.Errorf("error should mention 'inclde', got: %v", err)
	}
}

// --- Spec merging ---

func TestInclude_SpecsFromIncludedFileAvailable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "shared.yaml", `
specs:
  shared-spec:
    command:
      cli: echo shared
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./shared.yaml
specs:
  local-spec:
    command:
      cli: echo local
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := plan.Specs["local-spec"]; !ok {
		t.Error("local-spec should be present")
	}
	if _, ok := plan.Specs["shared-spec"]; !ok {
		t.Error("shared-spec from included file should be present")
	}
}

func TestInclude_BaseSpecOverridesIncluded(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "shared.yaml", `
specs:
  common:
    command:
      cli: echo from-shared
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./shared.yaml
specs:
  common:
    command:
      cli: echo from-main
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Specs["common"].Command.Cli != "echo from-main" {
		t.Errorf("base spec should override included; got cli=%q", plan.Specs["common"].Command.Cli)
	}
}

func TestInclude_SpecOrderBaseFirst(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "shared.yaml", `
specs:
  shared-a:
    command:
      cli: echo a
  shared-b:
    command:
      cli: echo b
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./shared.yaml
specs:
  main-spec:
    command:
      cli: echo main
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// main-spec should come first, then shared specs
	if len(plan.specOrder) != 3 {
		t.Fatalf("expected 3 specs in order, got %d: %v", len(plan.specOrder), plan.specOrder)
	}
	if plan.specOrder[0] != "main-spec" {
		t.Errorf("first spec should be main-spec, got %q", plan.specOrder[0])
	}
}

// --- Var merging ---

func TestInclude_VarsFromIncludedAvailable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `
vars:
  tool: ./bin/mytool
  env: staging
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./base.yaml
vars:
  env: production
specs:
  s:
    command:
      cli: echo
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// base vars become available
	if plan.Vars["tool"] != "./bin/mytool" {
		t.Errorf("vars[tool] = %q, want ./bin/mytool", plan.Vars["tool"])
	}
	// main file's vars take precedence
	if plan.Vars["env"] != "production" {
		t.Errorf("vars[env] = %q, want production (main should override included)", plan.Vars["env"])
	}
}

// --- Precondition merging ---

func TestInclude_PreconditionsMerged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `
preconditions:
  fs:
    - file: ./from-base
      exists: true
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./base.yaml
preconditions:
  fs:
    - file: ./from-main
      exists: true
specs:
  s:
    command:
      cli: echo
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fa := plan.Preconditions.FileSystemAssertions
	if len(fa) != 2 {
		t.Fatalf("expected 2 precondition assertions, got %d", len(fa))
	}
	// included preconditions are prepended
	if fa[0].File != "./from-base" {
		t.Errorf("first precondition should be from included file (from-base), got %q", fa[0].File)
	}
	if fa[1].File != "./from-main" {
		t.Errorf("second precondition should be from main file (from-main), got %q", fa[1].File)
	}
}

// --- Setup/teardown merging ---

func TestInclude_SetupMerged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `
setup:
  - cli: echo base-setup
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./base.yaml
setup:
  - cli: echo main-setup
specs:
  s:
    command:
      cli: echo
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Setup) != 2 {
		t.Fatalf("expected 2 setup commands, got %d", len(plan.Setup))
	}
	// included setup runs first
	if plan.Setup[0].Cli != "echo base-setup" {
		t.Errorf("included setup should run first, got %q", plan.Setup[0].Cli)
	}
}

func TestInclude_TeardownMerged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.yaml", `
teardown:
  - cli: echo base-teardown
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./base.yaml
teardown:
  - cli: echo main-teardown
specs:
  s:
    command:
      cli: echo
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Teardown) != 2 {
		t.Fatalf("expected 2 teardown commands, got %d", len(plan.Teardown))
	}
	// base teardown runs first, included teardown appended
	if plan.Teardown[0].Cli != "echo main-teardown" {
		t.Errorf("main teardown should run first, got %q", plan.Teardown[0].Cli)
	}
	if plan.Teardown[1].Cli != "echo base-teardown" {
		t.Errorf("included teardown should run last, got %q", plan.Teardown[1].Cli)
	}
}

// --- Transitive includes ---

func TestInclude_Transitive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "level2.yaml", `
specs:
  level2-spec:
    command:
      cli: echo level2
`)
	writeFile(t, dir, "level1.yaml", `
include:
  - ./level2.yaml
specs:
  level1-spec:
    command:
      cli: echo level1
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./level1.yaml
specs:
  main-spec:
    command:
      cli: echo main
`)
	plan, err := loadPlanFile(mainPath, make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"main-spec", "level1-spec", "level2-spec"} {
		if _, ok := plan.Specs[name]; !ok {
			t.Errorf("expected spec %q to be present after transitive include", name)
		}
	}
}

// --- Error cases ---

func TestInclude_MissingFile(t *testing.T) {
	dir := t.TempDir()
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./nonexistent.yaml
specs:
  s:
    command:
      cli: echo
`)
	_, err := loadPlanFile(mainPath, make(map[string]bool))
	if err == nil {
		t.Fatal("expected error for missing included file, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent.yaml") {
		t.Errorf("error should mention the missing file, got: %v", err)
	}
}

func TestInclude_CircularDetected(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.yaml")
	bPath := filepath.Join(dir, "b.yaml")
	writeFile(t, dir, "a.yaml", `
include:
  - ./b.yaml
specs:
  a:
    command:
      cli: echo a
`)
	writeFile(t, dir, "b.yaml", `
include:
  - ./a.yaml
specs:
  b:
    command:
      cli: echo b
`)
	_ = aPath
	_ = bPath
	_, err := loadPlanFile(aPath, make(map[string]bool))
	if err == nil {
		t.Fatal("expected error for circular include, got nil")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error should mention circular include, got: %v", err)
	}
}

// --- ExecuteFile integration ---

func TestExecuteFile_IncludedSpecsRun(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "shared.yaml", `
specs:
  shared:
    command:
      cli: echo shared
    expectations:
      status:
        equals_to: 0
`)
	mainPath := writeFile(t, dir, "main.yaml", `
include:
  - ./shared.yaml
specs:
  local:
    command:
      cli: echo local
    expectations:
      status:
        equals_to: 0
`)
	sut := NewLauncher()
	res, err := sut.ExecuteFile(mainPath)
	if err != nil {
		t.Fatalf("ExecuteFile: %v", err)
	}
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
	// both local and shared specs should have run
	blocks := res.Blocks()
	if len(blocks) != 2 {
		t.Errorf("expected 2 spec blocks, got %d", len(blocks))
	}
}
