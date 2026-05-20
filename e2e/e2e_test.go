//go:build !windows
// +build !windows

package e2e

import "testing"

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
