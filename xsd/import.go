package xsd

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveImports loads each xs:import directive's target schema and returns
// them as separate Schema instances. Unlike xs:include (which merges the
// included schema's content into the parent's target namespace), xs:import
// brings in a DIFFERENT namespace; the imported schemas must remain distinct
// so callers can preserve that separation (typically by appending them to
// the enclosing wsdl:definitions Types.Schemas slice).
//
// Each returned schema is fully resolved: its own xs:include directives and
// transitive xs:import directives are processed recursively. The returned
// slice is flat and contains every schema reachable via imports.
//
// selfPath is the path of the file (or containing WSDL file) that this
// schema was loaded from. It is used both to mark the root as visited (so
// a transitive import back to it is a no-op) and to resolve relative
// schemaLocation values.
//
// Schemas are deduped by absolute file path: an import reachable by
// multiple paths is loaded once.
//
// The xs:import spec permits omitting schemaLocation (leaving it to the
// processor to resolve the namespace via some external mechanism like an
// OASIS XML catalog). soap-go does not currently support that, but WSDLs
// in the wild sometimes emit duplicate <xs:import> elements for the same
// namespace where only one carries schemaLocation; we dedupe per namespace
// and use the resolvable variant. Imports whose entire group lacks a
// schemaLocation still return an error identifying the namespace.
//
// After a successful call, Imports is cleared.
func (s *Schema) ResolveImports(selfPath string) ([]Schema, error) {
	visited := map[string]bool{}
	if selfPath != "" {
		abs, err := filepath.Abs(selfPath)
		if err != nil {
			return nil, fmt.Errorf("resolve self path %q: %w", selfPath, err)
		}
		visited[abs] = true
	}
	return s.resolveImports(filepath.Dir(selfPath), visited)
}

// dedupeImportsByNamespace collapses multiple <xs:import> entries for the
// same namespace into a single entry, preferring one that carries a
// schemaLocation. First-encounter order is preserved across distinct
// namespaces for deterministic processing.
//
// Conflicting declarations — two imports of the same namespace with two
// different non-empty schemaLocations — return an error. soap-go will not
// guess which location is authoritative; silently dropping one risks losing
// types the WSDL references, and loading both produces duplicate schemas
// for the same namespace downstream. The comparison is on raw attribute
// strings, so "foo.xsd" and "./foo.xsd" would be flagged even though they
// resolve to the same file; we can add path-normalising comparison later
// if a real WSDL trips that.
func dedupeImportsByNamespace(imports []Import) ([]Import, error) {
	chosen := map[string]Import{}
	order := []string{}
	for _, imp := range imports {
		existing, seen := chosen[imp.Namespace]
		if !seen {
			chosen[imp.Namespace] = imp
			order = append(order, imp.Namespace)
			continue
		}
		// Conflicting non-empty locations — refuse to guess.
		if existing.SchemaLocation != "" && imp.SchemaLocation != "" &&
			existing.SchemaLocation != imp.SchemaLocation {
			return nil, fmt.Errorf(
				"xs:import of namespace %q declared with conflicting "+
					"schemaLocations %q and %q",
				imp.Namespace, existing.SchemaLocation, imp.SchemaLocation,
			)
		}
		// Upgrade to the variant that has a schemaLocation; otherwise
		// keep the first-encountered entry.
		if existing.SchemaLocation == "" && imp.SchemaLocation != "" {
			chosen[imp.Namespace] = imp
		}
	}
	result := make([]Import, 0, len(order))
	for _, ns := range order {
		result = append(result, chosen[ns])
	}
	return result, nil
}

func (s *Schema) resolveImports(baseDir string, visited map[string]bool) ([]Schema, error) {
	var result []Schema
	deduped, err := dedupeImportsByNamespace(s.Imports)
	if err != nil {
		return nil, err
	}
	for _, imp := range deduped {
		if imp.SchemaLocation == "" {
			return nil, fmt.Errorf(
				"xs:import of namespace %q has no schemaLocation; soap-go "+
					"does not currently resolve imports without schemaLocation",
				imp.Namespace,
			)
		}
		loc := imp.SchemaLocation
		if !filepath.IsAbs(loc) {
			loc = filepath.Join(baseDir, loc)
		}
		abs, err := filepath.Abs(loc)
		if err != nil {
			return nil, fmt.Errorf("xs:import %q: %w", imp.SchemaLocation, err)
		}
		if visited[abs] {
			continue
		}
		visited[abs] = true

		f, err := os.Open(abs)
		if err != nil {
			return nil, fmt.Errorf("xs:import %q: %w", imp.SchemaLocation, err)
		}
		imported, err := Parse(f)
		closeErr := f.Close()
		if err != nil {
			return nil, fmt.Errorf("xs:import %q: parse: %w", imp.SchemaLocation, err)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("xs:import %q: close: %w", imp.SchemaLocation, closeErr)
		}

		// Resolve the imported schema's own xs:include directives before
		// handing it back. xs:include merges into the same namespace, so
		// this mutates `imported` in place.
		if err := imported.resolveIncludes(filepath.Dir(abs), map[string]bool{abs: true}); err != nil {
			return nil, fmt.Errorf("xs:import %q: resolve includes: %w", imp.SchemaLocation, err)
		}

		// Recurse into the imported schema's own xs:import directives. Their
		// resolved schemas become siblings alongside this one in the result.
		transitive, err := imported.resolveImports(filepath.Dir(abs), visited)
		if err != nil {
			return nil, fmt.Errorf("xs:import %q: resolve transitive imports: %w", imp.SchemaLocation, err)
		}

		result = append(result, *imported)
		result = append(result, transitive...)
	}

	s.Imports = nil
	return result, nil
}
