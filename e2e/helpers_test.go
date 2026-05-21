package e2e

import (
	"testing"

	"github.com/enr/qac"
)

// run executes a YAML spec file and fails the test on any assertion error.
// The path is relative to the e2e/ directory (Go test CWD for this package).
func run(t *testing.T, specFile string) {
	t.Helper()
	launcher := qac.NewLauncher()
	report, err := launcher.ExecuteFile(specFile)
	if err != nil {
		t.Fatalf("ExecuteFile(%q): %v", specFile, err)
	}
	reporter := qac.NewTestLogsReporter(t)
	reporter.Publish(report)
	for _, err := range report.AllErrors() {
		t.Errorf("%v", err)
	}
}
