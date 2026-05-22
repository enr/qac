package qac

import (
	"strings"
	"testing"
)

// --- mergeEnv ---

func TestMergeEnv_CustomEntryPresent(t *testing.T) {
	result := mergeEnv(map[string]string{"QAC_UNIT_TEST_KEY_UNIQUE_XYZ": "hello"})
	for _, entry := range result {
		if entry == "QAC_UNIT_TEST_KEY_UNIQUE_XYZ=hello" {
			return
		}
	}
	t.Error("custom env entry not found in merged result")
}

func TestMergeEnv_DuplicateKeyCustomIsLast(t *testing.T) {
	// mergeEnv appends custom entries after os.Environ(); when the same key
	// appears twice the last occurrence wins on Linux and macOS.
	t.Setenv("QAC_DUP_KEY_TEST_UNIQUE", "parent")
	result := mergeEnv(map[string]string{"QAC_DUP_KEY_TEST_UNIQUE": "custom"})
	last := ""
	for _, entry := range result {
		if strings.HasPrefix(entry, "QAC_DUP_KEY_TEST_UNIQUE=") {
			last = entry
		}
	}
	if last != "QAC_DUP_KEY_TEST_UNIQUE=custom" {
		t.Errorf("expected custom value to be last entry for duplicate key, got %q", last)
	}
}

func TestMergeEnv_EmptyCustom(t *testing.T) {
	before := len(mergeEnv(map[string]string{}))
	after := len(mergeEnv(nil))
	// Both should equal len(os.Environ()); just verify no panic and same size.
	if before != after {
		t.Errorf("empty and nil custom maps should produce same-length result: %d vs %d", before, after)
	}
}
