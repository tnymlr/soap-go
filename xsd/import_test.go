package xsd_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justinclift-prvidr/soap-go/xsd"
)

func targetNamespaces(schemas []xsd.Schema) []string {
	ns := make([]string, 0, len(schemas))
	for _, s := range schemas {
		ns = append(ns, s.TargetNamespace)
	}
	return ns
}

func TestResolveImports_Transitive(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "imports", "parent.xsd")
	s := parseFile(t, path)

	imported, err := s.ResolveImports(path)
	if err != nil {
		t.Fatalf("ResolveImports: %v", err)
	}

	// parent.xsd imports child.xsd (ns=child); child.xsd imports
	// grandchild.xsd (ns=grandchild). Both should appear in the result,
	// each as a separate schema in its own namespace.
	want := map[string]bool{
		"http://example.com/child":      true,
		"http://example.com/grandchild": true,
	}
	got := map[string]bool{}
	for _, ns := range targetNamespaces(imported) {
		got[ns] = true
	}
	for ns := range want {
		if !got[ns] {
			t.Errorf("missing imported schema with targetNamespace %q; have: %v", ns, targetNamespaces(imported))
		}
	}

	// The parent schema's Imports slice must be cleared after resolution.
	if len(s.Imports) != 0 {
		t.Errorf("Imports not cleared after resolution: %+v", s.Imports)
	}
	// Each returned schema must also have its own Imports cleared.
	for _, inc := range imported {
		if len(inc.Imports) != 0 {
			t.Errorf("imported schema %q has leftover Imports: %+v", inc.TargetNamespace, inc.Imports)
		}
	}
}

func TestResolveImports_Cycle(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "imports", "cycle_a.xsd")
	s := parseFile(t, path)

	imported, err := s.ResolveImports(path)
	if err != nil {
		t.Fatalf("ResolveImports with cycle should not error, got: %v", err)
	}

	// cycle_a imports cycle_b; cycle_b imports cycle_a. Only cycle_b
	// should appear in the result — the back-edge to cycle_a is a no-op
	// because it's already visited.
	if len(imported) != 1 {
		t.Fatalf("expected 1 imported schema (cycle_b only), got %d: %v", len(imported), targetNamespaces(imported))
	}
	if imported[0].TargetNamespace != "http://example.com/cycle-b" {
		t.Errorf("expected cycle_b, got %q", imported[0].TargetNamespace)
	}
}

func TestResolveImports_MissingFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "imports", "missing_import.xsd")
	s := parseFile(t, path)

	_, err := s.ResolveImports(path)
	if err == nil {
		t.Fatal("expected error for missing import, got nil")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected error to wrap fs.ErrNotExist, got: %v", err)
	}
	if !strings.Contains(err.Error(), "does_not_exist.xsd") {
		t.Errorf("expected error to mention the missing file, got: %v", err)
	}
}

func TestResolveImports_NoSchemaLocation(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "imports", "no_location.xsd")
	s := parseFile(t, path)

	_, err := s.ResolveImports(path)
	if err == nil {
		t.Fatal("expected error for import without schemaLocation, got nil")
	}
	// The error must identify the namespace so a human can figure out what
	// to do about it (e.g. provide a catalog mapping or a schemaLocation).
	if !strings.Contains(err.Error(), "http://example.com/somewhere") {
		t.Errorf("expected error to identify the namespace, got: %v", err)
	}
	if !strings.Contains(err.Error(), "schemaLocation") {
		t.Errorf("expected error to mention schemaLocation, got: %v", err)
	}
}

func TestResolveImports_NoImports(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "imports", "grandchild.xsd")
	s := parseFile(t, path)

	imported, err := s.ResolveImports(path)
	if err != nil {
		t.Fatalf("ResolveImports: %v", err)
	}
	if len(imported) != 0 {
		t.Errorf("expected no imported schemas, got: %v", targetNamespaces(imported))
	}
}

// TestResolveImports_IncludeBubblesImports covers the case where the
// parent schema xs:includes a sibling which itself has xs:imports. The
// include-resolution pass bubbles those imports up to the parent; a
// subsequent ResolveImports pass must then load them.
func TestResolveImports_IncludeBubblesImports(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "imports", "include_with_import.xsd")
	s := parseFile(t, path)

	// xs:include first — this merges mixed_include.xsd into s and
	// bubbles mixed_include.xsd's xs:import on "bubbled" up into s.Imports.
	if err := s.ResolveIncludes(path); err != nil {
		t.Fatalf("ResolveIncludes: %v", err)
	}

	imported, err := s.ResolveImports(path)
	if err != nil {
		t.Fatalf("ResolveImports: %v", err)
	}

	// Both the direct import ("foreign") and the bubbled-up import
	// ("bubbled") should have been loaded.
	want := map[string]bool{
		"http://example.com/foreign": true,
		"http://example.com/bubbled": true,
	}
	got := map[string]bool{}
	for _, ns := range targetNamespaces(imported) {
		got[ns] = true
	}
	for ns := range want {
		if !got[ns] {
			t.Errorf("missing imported schema %q; have: %v", ns, targetNamespaces(imported))
		}
	}
}
