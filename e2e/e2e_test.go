//go:build !windows
// +build !windows

package e2e

import (
	"testing"

	"github.com/enr/qac"
)

// run executes a YAML spec file and fails the test on any assertion error.
// Paths are relative to the e2e/ directory (Go test CWD for this package).
func run(t *testing.T, specFile string) {
	t.Helper()
	launcher := qac.NewLauncher()
	report := launcher.ExecuteFile(specFile)
	reporter := qac.NewTestLogsReporter(t)
	reporter.Publish(report)
	for _, err := range report.AllErrors() {
		t.Errorf("%v", err)
	}
}

func TestE2EEcho(t *testing.T) {
	run(t, "echo.yaml")
}

func TestE2ECat(t *testing.T) {
	run(t, "cat.yaml")
}

func TestE2EWc(t *testing.T) {
	run(t, "wc.yaml")
}

func TestE2EPwd(t *testing.T) {
	run(t, "pwd.yaml")
}

func TestE2EExitCodes(t *testing.T) {
	run(t, "exit_codes.yaml")
}
