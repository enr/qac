package qac

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// verify checks the command output against the assertion's expected values.
func (a *OutputAssertion) verify(context planContext) AssertionResult {
	result := AssertionResult{
		description: fmt.Sprintf(`output %s for %s`, a.id, context.commandResult.execution),
	}
	out := context.commandResult.stdout
	if a.id == `stderr` {
		out = context.commandResult.stderr
	}
	out = strings.TrimSpace(out)
	shouldBeEmpty := a.IsEmpty != nil && *a.IsEmpty
	if shouldBeEmpty && out != "" {
		result.addErrorf("%s: expected empty but got\n%s", a.id, snippet(out))
	}

	if a.EqualsTo != "" {
		et := strings.TrimSpace(a.EqualsTo)
		if out != et {
			result.addErrorf("%s: actual\n%s\nnot equal to:\n%s", a.id, snippet(out), snippet(et))
		}
	}
	if a.EqualsToFile != "" {
		f, err := resolvePath(a.EqualsToFile, context)
		if err != nil {
			result.addInfraError(fmt.Errorf("resolving equals_to_file path %q: %w", a.EqualsToFile, err))
			return result
		}
		content, err := os.ReadFile(f)
		if err != nil {
			result.addInfraError(fmt.Errorf("reading equals_to_file %q: %w", f, err))
			return result
		}
		t := strings.TrimSpace(string(content))
		if out != t {
			result.addErrorf("%s: actual\n%s\nnot equal to:\n%s", a.id, snippet(out), snippet(t))
		}
	}
	if a.StartsWith != "" {
		if !strings.HasPrefix(out, a.StartsWith) {
			result.addErrorf("%s: actual output\n%s\ndoes not start with:\n%s", a.id, snippet(out), a.StartsWith)
		}
	}
	if a.EndsWith != "" {
		if !strings.HasSuffix(out, a.EndsWith) {
			result.addErrorf("%s: actual output\n%s\ndoes not end with:\n%s", a.id, snippetTail(out), a.EndsWith)
		}
	}
	if len(a.ContainsAll) > 0 {
		for _, t := range a.ContainsAll {
			if !strings.Contains(out, t) {
				result.addErrorf("%s: actual output\n%s\ndoes not contain:\n%s", a.id, snippet(out), t)
			}
		}
	}
	if len(a.ContainsAny) > 0 {
		if a.failContainsAny(out) {
			result.addErrorf("%s: actual output\n%s\ndoes not contain any of:\n%q", a.id, snippet(out), a.ContainsAny)
		}
	}
	if len(a.ContainsNone) > 0 {
		for _, t := range a.ContainsNone {
			if strings.Contains(out, t) {
				result.addErrorf("%s: actual output\n%s\ncontains:\n%s", a.id, snippet(out), t)
			}
		}
	}
	if a.Matches != "" {
		re, err := regexp.Compile(a.Matches)
		if err != nil {
			result.addConfigError(fmt.Errorf("%s: invalid regex in matches %q: %w", a.id, a.Matches, err))
		} else if !re.MatchString(out) {
			result.addErrorf("%s: actual output\n%s\ndoes not match:\n%s", a.id, snippet(out), a.Matches)
		}
	}
	if a.NotMatches != "" {
		re, err := regexp.Compile(a.NotMatches)
		if err != nil {
			result.addConfigError(fmt.Errorf("%s: invalid regex in not_matches %q: %w", a.id, a.NotMatches, err))
		} else if re.MatchString(out) {
			result.addErrorf("%s: actual output\n%s\nshould not match:\n%s", a.id, snippet(out), a.NotMatches)
		}
	}

	return result
}

func (a *OutputAssertion) failContainsAny(out string) bool {
	fail := true
	for _, t := range a.ContainsAny {
		if strings.Contains(out, t) {
			fail = false
			break
		}
	}
	return fail
}
