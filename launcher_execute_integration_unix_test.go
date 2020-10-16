// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package qac

func testFiles() []string {
	return []string{
		`examples/test-ok-linux.yaml`,
		`examples/linux.yaml`,
	}
}
