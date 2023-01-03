//go:build windows
// +build windows

package qac

import (
	"testing"
)

type windowsPathTestCase struct {
	basedir  string
	declared string
	expected string
}

var windowsPaths = []windowsPathTestCase{
	{
		basedir:  `C:\projects\`,
		declared: `.\..\aaa`,
		expected: `C:\aaa`,
	},
	{
		basedir:  `C:\projects\test`,
		declared: `..`,
		expected: `C:\projects`,
	},
	{
		basedir:  `C:/projects/`,
		declared: `./../aaa`,
		expected: `C:\aaa`,
	},
	{
		basedir:  `C:/projects/test`,
		declared: `..`,
		expected: `C:\projects`,
	},
}

func TestResolveWindowsPath(t *testing.T) {
	for _, p := range windowsPaths {
		context := planContext{
			basedir: p.basedir,
		}
		actual, _ := resolvePath(p.declared, context)
		if actual != p.expected {
			t.Errorf(`Windows path resolution error: expected %s got %s`, p.expected, actual)
		}
	}
}
