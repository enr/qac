package qac

import (
	"testing"
)

func TestFileSystemAssertionFactory_BothFileAndDirectory(t *testing.T) {
	a := &FileSystemAssertion{File: "foo.txt", Directory: "dir"}
	_, err := a.actualAssertion(planContext{})
	if err == nil {
		t.Fatal("expected error when both file and directory are set")
	}
}

func TestFileSystemAssertionFactory_NeitherSet_ReturnsDirectory(t *testing.T) {
	// When neither file nor directory is set the factory falls through to the
	// directory branch and returns a DirectoryAssertion with an empty path.
	a := &FileSystemAssertion{}
	result, err := a.actualAssertion(planContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(*DirectoryAssertion); !ok {
		t.Errorf("expected *DirectoryAssertion, got %T", result)
	}
}

func TestFileSystemAssertionFactory_ExistsNil_DefaultsToTrue(t *testing.T) {
	a := &FileSystemAssertion{File: "foo.txt"} // Exists is nil
	result, err := a.actualAssertion(planContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fa, ok := result.(*FileAssertion)
	if !ok {
		t.Fatalf("expected *FileAssertion, got %T", result)
	}
	if !fa.Exists {
		t.Error("expected Exists to default to true when the YAML field is absent")
	}
}

func TestFileSystemAssertionFactory_ExistsExplicitFalse(t *testing.T) {
	b := false
	a := &FileSystemAssertion{File: "foo.txt", Exists: &b}
	result, err := a.actualAssertion(planContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fa, ok := result.(*FileAssertion)
	if !ok {
		t.Fatalf("expected *FileAssertion, got %T", result)
	}
	if fa.Exists {
		t.Error("expected Exists=false to be propagated to the FileAssertion")
	}
}

func TestFileSystemAssertionFactory_FieldsPropagate_File(t *testing.T) {
	a := &FileSystemAssertion{
		File:         "f.txt",
		EqualsTo:     "ref.txt",
		TextEqualsTo: "ref2.txt",
		ContainsAll:  []string{"aa"},
		ContainsAny:  []string{"bb"},
	}
	result, err := a.actualAssertion(planContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fa := result.(*FileAssertion)
	if fa.EqualsTo != "ref.txt" || fa.TextEqualsTo != "ref2.txt" ||
		len(fa.ContainsAll) != 1 || len(fa.ContainsAny) != 1 {
		t.Errorf("file assertion fields not propagated correctly: %+v", fa)
	}
}

func TestFileSystemAssertionFactory_FieldsPropagate_Directory(t *testing.T) {
	a := &FileSystemAssertion{
		Directory:       "d",
		EqualsTo:        "ref",
		ContainsAll:     []string{"a"},
		ContainsAny:     []string{"b"},
		ContainsExactly: []string{"c"},
	}
	result, err := a.actualAssertion(planContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	da := result.(*DirectoryAssertion)
	if da.EqualsTo != "ref" || len(da.ContainsAll) != 1 ||
		len(da.ContainsAny) != 1 || len(da.ContainsExactly) != 1 {
		t.Errorf("directory assertion fields not propagated correctly: %+v", da)
	}
}
