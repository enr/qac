package qac

import (
	"strings"
	"testing"
)

func TestSnippetShortPassThrough(t *testing.T) {
	s := "short string"
	if snippet(s) != s {
		t.Errorf("expected short string to pass through unchanged")
	}
}

func TestSnippetLongTruncated(t *testing.T) {
	long := strings.Repeat("x", snippetMaxLen+1)
	out := snippet(long)
	if len(out) <= snippetMaxLen {
		// the truncation hint is appended, so the result is longer than snippetMaxLen
		// but the raw content portion is exactly snippetMaxLen
		t.Errorf("expected truncated output to be longer than snippetMaxLen due to hint")
	}
	if !strings.Contains(out, "QAC_VERBOSE=1") {
		t.Error("expected truncation hint to mention QAC_VERBOSE=1")
	}
	if !strings.HasPrefix(out, long[:snippetMaxLen]) {
		t.Error("expected truncated output to start with first snippetMaxLen bytes")
	}
}

func TestSnippetTailShortPassThrough(t *testing.T) {
	s := "short"
	if snippetTail(s) != s {
		t.Errorf("expected short string to pass through unchanged")
	}
}

func TestSnippetTailLongShowsTail(t *testing.T) {
	prefix := strings.Repeat("a", snippetMaxLen+10)
	suffix := strings.Repeat("z", snippetMaxLen)
	long := prefix + suffix
	out := snippetTail(long)
	if !strings.HasSuffix(out, suffix) {
		t.Error("expected tail of output to match last snippetMaxLen bytes of input")
	}
	if !strings.Contains(out, "last") {
		t.Error("expected truncation hint mentioning 'last'")
	}
}
