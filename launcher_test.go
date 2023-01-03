package qac

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fixedValueExecutor struct {
	success  bool
	exitCode int
	stdout   string
	stderr   string
}

func (e *fixedValueExecutor) execute(c Command) executionResult {
	return executionResult{
		success:  e.success,
		exitCode: e.exitCode,
		stdout:   e.stdout,
		stderr:   e.stderr,
	}
}

func randomExitCode() int {
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := 128
	return (rand.Intn(max-min+1) + min)
}

func ignoreTestSpecificationError(t *testing.T) {
	stdout := fmt.Sprintf(`stdout-%d`, time.Now().UnixNano())
	stderr := fmt.Sprintf(`stderr-%d`, time.Now().UnixNano())

	e := &fixedValueExecutor{
		success:  true,
		exitCode: randomExitCode(),
		stdout:   stdout,
		stderr:   stderr,
	}
	sut := newLauncher(e)

	/*
	 *  A Spec with wrong assertions.
	 */
	expectations := Expectations{
		StatusAssertion: StatusAssertion{
			EqualsTo: `0`,
		},
		OutputAssertions: OutputAssertions{
			Stdout: OutputAssertion{
				EqualsTo: `wrong-stdout`,
			},
			Stderr: OutputAssertion{
				EqualsTo: `wrong-stderr`,
			},
		},
	}

	spec := Spec{
		Command: Command{
			Exe:  "test",
			Args: []string{},
		},
		Expectations: expectations,
	}

	specs := make(map[string]Spec)
	specs[`test1`] = spec
	plan := TestPlan{
		Specs: specs,
	}

	res := sut.Execute(plan)

	reporter := NewConsoleReporter()
	reporter.Publish(res)

	if !atLeastOneErrorContaining(res.AllErrors(), "wrong-stdout") {
		t.Errorf("Expected at least one error containing <%s>", "wrong-stdout")
	}
	if !atLeastOneErrorContaining(res.AllErrors(), "wrong-stderr") {
		t.Errorf("Expected at least one error containing <%s>", "wrong-stderr")
	}
	if len(res.AllErrors()) != 3 {
		t.Errorf(`Expected 3 errors, got %d`, len(res.AllErrors()))
	}
}

func atLeastOneErrorContaining(errors []error, expected string) bool {
	for _, err := range errors {
		if strings.Contains(err.Error(), expected) {
			return true
		}
	}
	return false
}

func TestSpecificationOk(t *testing.T) {
	stdout := fmt.Sprintf(`stdout-%d`, time.Now().UnixNano())
	stderr := fmt.Sprintf(`stderr-%d`, time.Now().UnixNano())
	exitCode := randomExitCode()

	e := &fixedValueExecutor{
		success:  true,
		exitCode: exitCode,
		stdout:   stdout,
		stderr:   stderr,
	}
	sut := newLauncher(e)

	/*
	 *  A Spec with no errors.
	 */
	expectations := Expectations{
		StatusAssertion: StatusAssertion{
			EqualsTo: strconv.Itoa(exitCode),
		},
		OutputAssertions: OutputAssertions{
			Stdout: OutputAssertion{
				EqualsTo: stdout,
			},
			Stderr: OutputAssertion{
				EqualsTo: stderr,
			},
		},
	}

	spec := Spec{
		Command: Command{
			Exe:  "test",
			Args: []string{},
		},
		Expectations: expectations,
	}

	specs := make(map[string]Spec)
	specs[`test1`] = spec
	plan := TestPlan{
		Specs: specs,
	}

	res := sut.Execute(plan)

	reporter := NewConsoleReporter()
	reporter.Publish(res)

	errors := res.AllErrors()
	if len(errors) > 0 {
		t.Errorf(`expected 0 errors but got %d`, len(errors))
		for ei, err := range errors {
			t.Errorf(`%d error %v`, ei, err)
		}
	}

}

func TestOutputContainsNone(t *testing.T) {
	stdout := fmt.Sprintf(`stdout-%d`, time.Now().UnixNano())
	stderr := fmt.Sprintf(`stderr-%d`, time.Now().UnixNano())
	exitCode := randomExitCode()

	e := &fixedValueExecutor{
		success:  true,
		exitCode: exitCode,
		stdout:   stdout,
		stderr:   stderr,
	}
	sut := newLauncher(e)

	/*
	 *  A Spec with no errors.
	 */
	expectations := Expectations{
		StatusAssertion: StatusAssertion{
			EqualsTo: strconv.Itoa(exitCode),
		},
		OutputAssertions: OutputAssertions{
			Stdout: OutputAssertion{
				ContainsNone: []string{stdout},
			},
			Stderr: OutputAssertion{
				EqualsTo: stderr,
			},
		},
	}

	spec := Spec{
		Command: Command{
			Exe:  "test",
			Args: []string{},
		},
		Expectations: expectations,
	}

	specs := make(map[string]Spec)
	specs[`test1`] = spec
	plan := TestPlan{
		Specs: specs,
	}

	res := sut.Execute(plan)

	reporter := NewConsoleReporter()
	reporter.Publish(res)

	errors := res.AllErrors()
	if len(errors) != 1 {
		t.Errorf(`expected 1 errors but got %d`, len(errors))
	}
	if len(errors) == 1 {
		ae := errors[0]
		if !strings.Contains(ae.Error(), stdout) {
			t.Errorf(`error does not contain "%s" but: "%s"`, stdout, ae.Error())
		}
	}

}
