package qac

import (
	"os"
	"testing"
)

func TestInterpolateNoVars(t *testing.T) {
	src := []byte("cli: mytool --out result.txt")
	got := interpolate(src, nil)
	if string(got) != string(src) {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestInterpolatePlanVars(t *testing.T) {
	vars := map[string]string{
		"tool": "./bin/mytool",
		"base": "/tmp/workdir",
	}
	src := []byte("cli: ${tool} create ${base}/output")
	got := string(interpolate(src, vars))
	want := "cli: ./bin/mytool create /tmp/workdir/output"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInterpolateEnvVar(t *testing.T) {
	os.Setenv("QAC_TEST_VAR", "hello")
	defer os.Unsetenv("QAC_TEST_VAR")

	src := []byte("cli: echo ${env.QAC_TEST_VAR}")
	got := string(interpolate(src, nil))
	want := "cli: echo hello"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInterpolateUnknownPlaceholderUnchanged(t *testing.T) {
	src := []byte("cli: ${unknown} does-not-exist")
	got := string(interpolate(src, nil))
	if got != string(src) {
		t.Errorf("unknown placeholder should be left unchanged; got %q", got)
	}
}

func TestInterpolatePlanVarWithEnvRef(t *testing.T) {
	os.Setenv("QAC_TEST_HOME", "/home/testuser")
	defer os.Unsetenv("QAC_TEST_HOME")

	// Plan var value contains an env reference; the env pass resolves it.
	vars := map[string]string{
		"base": "${env.QAC_TEST_HOME}/workdir",
	}
	src := []byte("cli: mytool --dir ${base}")
	got := string(interpolate(src, vars))
	want := "cli: mytool --dir /home/testuser/workdir"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
