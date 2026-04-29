package xsd_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tnymlr/soap-go/xsd"
)

func parseFile(t *testing.T, path string) *xsd.Schema {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	t.Cleanup(func() { _ = f.Close() })
	s, err := xsd.Parse(f)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return s
}

func elementNames(s *xsd.Schema) []string {
	names := make([]string, 0, len(s.Elements))
	for _, e := range s.Elements {
		names = append(names, e.Name)
	}
	return names
}

func TestResolveIncludes_Nested(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "includes", "parent.xsd")
	s := parseFile(t, path)

	if err := s.ResolveIncludes(path); err != nil {
		t.Fatalf("ResolveIncludes: %v", err)
	}

	// Content from parent, both children, and the transitive grandchild
	// should all appear as top-level elements on the merged schema.
	want := map[string]bool{
		"ParentElement":     true,
		"Child1Element":     true,
		"Child2Element":     true,
		"GrandchildElement": true,
	}
	got := map[string]bool{}
	for _, name := range elementNames(s) {
		got[name] = true
	}
	for name := range want {
		if !got[name] {
			t.Errorf("missing merged element %q; have: %v", name, elementNames(s))
		}
	}

	// Complex and simple types from included schemas should also be merged.
	if s.ResolveComplexType("Child1Type") == nil {
		t.Errorf("Child1Type not merged from child1.xsd")
	}
	if s.ResolveSimpleType("Child2Type") == nil {
		t.Errorf("Child2Type not merged from child2.xsd")
	}

	// Includes should be cleared after successful resolution.
	if len(s.Includes) != 0 {
		t.Errorf("Includes not cleared after resolution: %+v", s.Includes)
	}
}

func TestResolveIncludes_Cycle(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "includes", "cycle_a.xsd")
	s := parseFile(t, path)

	if err := s.ResolveIncludes(path); err != nil {
		t.Fatalf("ResolveIncludes with cycle should return nil, got: %v", err)
	}

	// Both elements should be present exactly once; no infinite recursion,
	// no duplicates from revisiting.
	counts := map[string]int{}
	for _, e := range s.Elements {
		counts[e.Name]++
	}
	if counts["AElement"] != 1 {
		t.Errorf("expected AElement once, got %d", counts["AElement"])
	}
	if counts["BElement"] != 1 {
		t.Errorf("expected BElement once, got %d", counts["BElement"])
	}
}

func TestResolveIncludes_MissingFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "includes", "missing_include.xsd")
	s := parseFile(t, path)

	err := s.ResolveIncludes(path)
	if err == nil {
		t.Fatal("expected error for missing include, got nil")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected error to wrap fs.ErrNotExist, got: %v", err)
	}
	if !strings.Contains(err.Error(), "does_not_exist.xsd") {
		t.Errorf("expected error to mention the missing file, got: %v", err)
	}
}

func TestResolveIncludes_NoIncludes(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "includes", "child2.xsd")
	s := parseFile(t, path)

	before := len(s.Elements)
	if err := s.ResolveIncludes(path); err != nil {
		t.Fatalf("ResolveIncludes: %v", err)
	}
	if got := len(s.Elements); got != before {
		t.Errorf("elements count changed from %d to %d on a schema with no includes", before, got)
	}
}
