package qac

import (
	"testing"
)

type statusAssertionTestCase struct {
	exitCode        int
	value           int
	expectedSuccess bool
}

var statusEqualsToSpecs = []statusAssertionTestCase{
	{
		exitCode:        1,
		value:           1,
		expectedSuccess: true,
	},
	{
		exitCode:        0,
		value:           1,
		expectedSuccess: false,
	},
	{
		exitCode:        1,
		value:           0,
		expectedSuccess: false,
	},
}

func TestIsEqualAssertion(t *testing.T) {

	for _, spec := range statusEqualsToSpecs {

		context := planContext{
			commandResult: executionResult{
				exitCode: spec.exitCode,
			},
		}
		sut := &StatusAssertion{
			EqualsTo: &spec.value,
		}
		assertionResult := sut.verify(context)
		if assertionResult.Success() != spec.expectedSuccess {
			t.Errorf(`status assertion expected %t but got %t for exit code %d (expected equals to %d)`, spec.expectedSuccess, assertionResult.Success(), spec.exitCode, spec.value)
		}
	}
}

var statusGreaterThenSpecs = []statusAssertionTestCase{
	{
		exitCode:        1,
		value:           0,
		expectedSuccess: true,
	},
	{
		exitCode:        1,
		value:           1,
		expectedSuccess: false,
	},
	{
		exitCode:        1,
		value:           2,
		expectedSuccess: false,
	},
}

func TestGreaterThanAssertion(t *testing.T) {

	for _, spec := range statusGreaterThenSpecs {

		context := planContext{
			commandResult: executionResult{
				exitCode: spec.exitCode,
			},
		}
		sut := &StatusAssertion{
			GreaterThan: &spec.value,
		}
		assertionResult := sut.verify(context)
		if assertionResult.Success() != spec.expectedSuccess {
			t.Errorf(`status assertion expected %t but got %t for exit code %d (expected greater than %d)`, spec.expectedSuccess, assertionResult.Success(), spec.exitCode, spec.value)
		}
	}

}

var statusLessThenSpecs = []statusAssertionTestCase{
	{
		exitCode:        0,
		value:           1,
		expectedSuccess: true,
	},
}

func TestLessThanAssertion(t *testing.T) {

	for _, spec := range statusLessThenSpecs {

		context := planContext{
			commandResult: executionResult{
				exitCode: spec.exitCode,
			},
		}
		sut := &StatusAssertion{
			LessThan: &spec.value,
		}
		assertionResult := sut.verify(context)
		if assertionResult.Success() != spec.expectedSuccess {
			t.Errorf(`status assertion expected %t but got %t for exit code %d (expected less than %d)`, spec.expectedSuccess, assertionResult.Success(), spec.exitCode, spec.value)
		}
	}

}
