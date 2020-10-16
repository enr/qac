package qac

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLauncherExecution(t *testing.T) {
	// prepare test filesystem
	err := os.RemoveAll("../../qac/tmp")
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = os.MkdirAll("../../qac/tmp", 0755)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = os.MkdirAll("../../qac/tmp/test-rm_r", 0755)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = createEmptyFile(`../../qac/tmp/test-rm_r/file`)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	err = createEmptyFile(`../../qac/tmp/test-rm`)
	if err != nil {
		t.Fatalf(`error prepare %v`, err)
	}
	files := testFiles()
	for _, testFile := range files {
		launcher := NewLauncher()
		report := launcher.ExecuteFile(testFile)
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
	return ioutil.WriteFile(name, d, 0644)
}
