package qac

import (
	"strings"
	"testing"
)

// --- prefixOverlap ---

func TestPrefixOverlap_FullMatch(t *testing.T) {
	if n := prefixOverlap("Status: OK", "line: Status: OK done"); n != 10 {
		t.Errorf("expected 10, got %d", n)
	}
}

func TestPrefixOverlap_PartialMatch(t *testing.T) {
	if n := prefixOverlap("Status: OK", "Status: FAIL"); n != 8 {
		t.Errorf("expected 8 (common prefix 'Status: '), got %d", n)
	}
}

func TestPrefixOverlap_NoMatch(t *testing.T) {
	if n := prefixOverlap("Status: OK", "completely different"); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

// --- contextHint ---

func TestContextHint_EmptyText(t *testing.T) {
	out := contextHint("", "anything")
	if !strings.Contains(out, "empty") {
		t.Errorf("expected '(empty)' for empty text, got %q", out)
	}
}

func TestContextHint_ShowsLineNumbers(t *testing.T) {
	text := "line one\nline two\nline three"
	out := contextHint(text, "line two")
	if !strings.Contains(out, "1:") || !strings.Contains(out, "2:") || !strings.Contains(out, "3:") {
		t.Errorf("expected line numbers in context output, got:\n%s", out)
	}
}

func TestContextHint_CentersOnBestMatch(t *testing.T) {
	// Target partially matches line 5; context should include it.
	var lines []string
	for i := 0; i < 10; i++ {
		lines = append(lines, "unrelated line")
	}
	lines[4] = "Status: FAIL"
	text := strings.Join(lines, "\n")
	out := contextHint(text, "Status: OK")
	if !strings.Contains(out, "5:") {
		t.Errorf("expected context centered around line 5 (partial match), got:\n%s", out)
	}
}

func TestContextHint_NoMatchDefaultsToStart(t *testing.T) {
	text := "first\nsecond\nthird\nfourth\nfifth\nsixth\nseventh\neighth"
	out := contextHint(text, "zzz-no-match-zzz")
	// With no partial match, bestMatchLine returns 0, so first lines should appear.
	if !strings.Contains(out, "1:") {
		t.Errorf("expected line 1 in context when no match found, got:\n%s", out)
	}
}

func TestContextHint_ShowsContextHeader(t *testing.T) {
	text := "line one\nline two"
	out := contextHint(text, "missing")
	if !strings.Contains(out, "context (lines") {
		t.Errorf("expected context header, got: %q", out)
	}
}

func TestContextHint_TruncatesLongLines(t *testing.T) {
	long := strings.Repeat("x", 200)
	out := contextHint(long, "")
	if strings.Contains(out, strings.Repeat("x", 200)) {
		t.Error("expected long line to be truncated in context output")
	}
	if !strings.Contains(out, "...") {
		t.Error("expected truncation marker '...' for long line")
	}
}

// --- contextHintAround ---

func TestContextHintAround_ShowsMatchLine(t *testing.T) {
	text := "alpha\nbeta\ngamma\ndelta\nfound it here\nepsilon"
	out := contextHintAround(text, "found it here")
	if !strings.Contains(out, "5:") {
		t.Errorf("expected context centered around line 5 where match occurs, got:\n%s", out)
	}
}

func TestContextHintAround_EmptyText(t *testing.T) {
	out := contextHintAround("", "anything")
	if !strings.Contains(out, "empty") {
		t.Errorf("expected '(empty)' for empty text, got %q", out)
	}
}

// --- contextHintTail ---

func TestContextHintTail_ShowsLastLines(t *testing.T) {
	var lines []string
	for i := 1; i <= 20; i++ {
		lines = append(lines, "line content")
	}
	text := strings.Join(lines, "\n")
	out := contextHintTail(text)
	if !strings.Contains(out, "20:") {
		t.Errorf("expected last line (20) in tail context, got:\n%s", out)
	}
}

func TestContextHintTail_EmptyText(t *testing.T) {
	out := contextHintTail("")
	if !strings.Contains(out, "empty") {
		t.Errorf("expected '(empty)' for empty text, got %q", out)
	}
}

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
