package qac

import (
	"os"
	"testing"
)

func TestLauncherExecution(t *testing.T) {
	// prepare test filesystem
	err := os.RemoveAll("./workdir")
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = os.MkdirAll("./workdir", 0755)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = os.MkdirAll("./workdir/test_rm_r/dir", 0755)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = createEmptyFile(`./workdir/test_rm_r/dir/file`)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = createEmptyFile(`./workdir/test_rm`)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	files := testFiles()
	for _, testFile := range files {
		launcher := NewLauncher()
		report, err := launcher.ExecuteFile(testFile)
		if err != nil {
			t.Errorf("ExecuteFile(%q): %v", testFile, err)
			continue
		}
		errors := report.AllErrors()
		if len(errors) > 0 {
			t.Errorf(`File %s: expected 0 errors but got %d: %q`, testFile, len(errors), errors)
		}
		reporter := NewConsoleReporter()
		reporter.Publish(report)
	}
}

func createEmptyFile(name string) error {
	d := []byte("")
	return os.WriteFile(name, d, 0644)
}
