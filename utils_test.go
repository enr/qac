// +build windows

package qac

import (
	"fmt"
	"testing"
)

func TestResolvePath1(t *testing.T) {
	declared := `.\..\aaa`
	context := planContext{
		basedir: `C:\projects\`,
	}
	actual, _ := resolvePath(declared, context)
	fmt.Printf("ACTUAL %s \n", actual)
}
func TestResolvePath2(t *testing.T) {
	declared := `..`
	context := planContext{
		basedir: `C:\projects\test`,
	}
	actual, _ := resolvePath(declared, context)
	fmt.Printf("ACTUAL %s \n", actual)
}
