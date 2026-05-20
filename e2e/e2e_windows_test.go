//go:build windows
// +build windows

package e2e

import "testing"

func TestE2EEcho(t *testing.T) {
	run(t, "echo_windows.yaml")
}

func TestE2EType(t *testing.T) {
	run(t, "type_windows.yaml")
}

func TestE2EDir(t *testing.T) {
	run(t, "dir_windows.yaml")
}

func TestE2ECd(t *testing.T) {
	run(t, "cd_windows.yaml")
}

func TestE2EExitCodes(t *testing.T) {
	run(t, "exit_codes_windows.yaml")
}
