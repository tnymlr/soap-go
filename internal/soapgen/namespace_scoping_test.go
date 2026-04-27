package soapgen

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/tnymlr/soap-go/wsdl"
	"github.com/tnymlr/soap-go/xsd"
)

// nestedInlineWSDL is a minimal WSDL containing three levels of inline
// anonymous complex types — Outer's complexType contains Middle as an
// inline child, Middle's complexType contains Inner as an inline child,
// Inner's complexType holds a leaf string field. Used to exercise
// register-pass and generate-pass name agreement at depth ≥ 2.
const nestedInlineWSDL = `<?xml version="1.0"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
    xmlns:tns="http://example.com/ns/v1"
    xmlns:xs="http://www.w3.org/2001/XMLSchema"
    targetNamespace="http://example.com/ns/v1">
  <types>
    <xs:schema targetNamespace="http://example.com/ns/v1" elementFormDefault="qualified">
      <xs:element name="Outer">
        <xs:complexType>
          <xs:sequence>
            <xs:element name="Middle">
              <xs:complexType>
                <xs:sequence>
                  <xs:element name="Inner">
                    <xs:complexType>
                      <xs:sequence>
                        <xs:element name="Value" type="xs:string"/>
                      </xs:sequence>
                    </xs:complexType>
                  </xs:element>
                </xs:sequence>
              </xs:complexType>
            </xs:element>
          </xs:sequence>
        </xs:complexType>
      </xs:element>
    </xs:schema>
  </types>
</definitions>`

// generateTypesContent runs the full code-generation pipeline against the
// given WSDL string and returns the contents of types.go.
func generateTypesContent(t *testing.T, wsdlXML string, cfg Config) string {
	t.Helper()

	var defs wsdl.Definitions
	if err := xml.Unmarshal([]byte(wsdlXML), &defs); err != nil {
		t.Fatalf("parse WSDL: %v", err)
	}

	g := NewGenerator(&defs, cfg)
	if err := g.Generate(); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, file := range g.Files() {
		if file.Filename() == "types.go" {
			content, err := file.Content()
			if err != nil {
				t.Fatalf("file.Content: %v", err)
			}
			return string(content)
		}
	}
	t.Fatalf("types.go was not generated")
	return ""
}

// TestNestedInlineTypes_NoPrefix is the baseline. Without namespace
// prefixing the register pass and the generate pass produce identical
// names, so parent-to-child references resolve cleanly.
func TestNestedInlineTypes_NoPrefix(t *testing.T) {
	t.Parallel()

	got := generateTypesContent(t, nestedInlineWSDL, Config{
		PackageName: "test",
	})

	for _, want := range []string{
		"type OuterWrapper struct",
		"type Outer_Middle struct",
		"type OuterMiddle_Inner struct",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in generated types.go, not found:\n%s", want, got)
		}
	}

	// No field should resolve to RawXML — every inline child has a
	// typed counterpart that the lookup must find.
	if strings.Contains(got, "*RawXML") || strings.Contains(got, "[]RawXML") {
		t.Errorf("unexpected RawXML field in generated types.go (no-prefix baseline):\n%s", got)
	}
}

// TestNestedInlineTypes_WithPrefix verifies that when NamespacePrefixes
// is configured, the register pass applies the prefix the same way the
// generate pass does. If the two passes drift, parent-to-child references
// at depth ≥ 2 fail their lookup and fall back to RawXML, even though the
// inline type was emitted correctly under the prefixed name.
func TestNestedInlineTypes_WithPrefix(t *testing.T) {
	t.Parallel()

	got := generateTypesContent(t, nestedInlineWSDL, Config{
		PackageName: "test",
		NamespacePrefixes: map[string]string{
			"http://example.com/ns/v1": "NS",
		},
	})

	// The expected inline-type names follow the same Outer_Inner +
	// prefix convention used by the generate pass at every nesting
	// depth. If these aren't present, the generator did not emit the
	// types under prefixed names.
	for _, want := range []string{
		"type NS_OuterWrapper struct",
		"type NS_Outer_Middle struct",
		"type NS_NSOuterMiddle_Inner struct",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in generated types.go, not found:\n%s", want, got)
		}
	}

	// Parent fields that should reference the prefixed inline-type
	// names must not fall back to RawXML. If this assertion ever
	// fails, the register pass has drifted from the generate pass on
	// prefix application again.
	if strings.Contains(got, "*RawXML") || strings.Contains(got, "[]RawXML") {
		t.Errorf("unexpected RawXML field in generated types.go (prefixed):\n%s", got)
	}
}

// TestXSDDateTime_Codegen verifies the generator emits soap.XSDDateTime
// references (not time.Time) for xs:dateTime fields, both as element text
// and as an attribute, and that the soap-go import is registered.
func TestXSDDateTime_Codegen(t *testing.T) {
	t.Parallel()

	const dateTimeWSDL = `<?xml version="1.0"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
    xmlns:tns="http://example.com/ns/v1"
    xmlns:xs="http://www.w3.org/2001/XMLSchema"
    targetNamespace="http://example.com/ns/v1">
  <types>
    <xs:schema targetNamespace="http://example.com/ns/v1" elementFormDefault="qualified">
      <xs:element name="Record">
        <xs:complexType>
          <xs:sequence>
            <xs:element name="When" type="xs:dateTime"/>
          </xs:sequence>
          <xs:attribute name="timestamp" type="xs:dateTime"/>
        </xs:complexType>
      </xs:element>
    </xs:schema>
  </types>
</definitions>`

	got := generateTypesContent(t, dateTimeWSDL, Config{
		PackageName: "test",
	})

	for _, want := range []string{
		"soap.XSDDateTime",
		`"github.com/justinclift-prvidr/soap-go"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in generated types.go, not found:\n%s", want, got)
		}
	}

	if strings.Contains(got, "time.Time") {
		t.Errorf("xs:dateTime field should not map to time.Time:\n%s", got)
	}
}

func TestNsPrefixedName(t *testing.T) {
	t.Parallel()

	g := NewGenerator(nil, Config{
		NamespacePrefixes: map[string]string{
			"http://example.com/core/v1":    "Core",
			"http://example.com/billing/v1": "CB",
		},
	})

	tests := []struct {
		ns   string
		name string
		want string
	}{
		{"http://example.com/core/v1", "FlexAttr", "Core_FlexAttr"},
		{"http://example.com/billing/v1", "Invoice", "CB_Invoice"},
		{"http://example.com/unknown/v1", "Foo", "Foo"}, // unmapped namespace
		{"", "Bar", "Bar"}, // empty namespace
	}

	for _, tt := range tests {
		got := g.nsPrefixedName(tt.ns, tt.name)
		if got != tt.want {
			t.Errorf("nsPrefixedName(%q, %q) = %q, want %q", tt.ns, tt.name, got, tt.want)
		}
	}
}

func TestNsPrefixedName_Disabled(t *testing.T) {
	t.Parallel()

	g := NewGenerator(nil, Config{})

	got := g.nsPrefixedName("http://example.com/core/v1", "FlexAttr")
	if got != "FlexAttr" {
		t.Errorf("expected plain name when scoping disabled, got %q", got)
	}
}

func TestNamespaceScopingEnabled(t *testing.T) {
	t.Parallel()

	g1 := NewGenerator(nil, Config{})
	if g1.namespaceScopingEnabled() {
		t.Error("expected false when no prefixes configured")
	}

	g2 := NewGenerator(nil, Config{
		NamespacePrefixes: map[string]string{"http://example.com": "EX"},
	})
	if !g2.namespaceScopingEnabled() {
		t.Error("expected true when prefixes configured")
	}
}

func TestResolveNsScopedGoName(t *testing.T) {
	t.Parallel()

	schema := &xsd.Schema{
		TargetNamespace: "http://example.com/service/v1",
	}
	// ExtraAttrs would normally be populated by XML parsing; not needed for these tests
	// since we're testing the generator/context methods directly

	g := NewGenerator(nil, Config{
		NamespacePrefixes: map[string]string{
			"http://example.com/service/v1": "SVC",
			"http://example.com/core/v1":    "Core",
		},
	})

	ctx := newSchemaContext(schema, g)

	// Test scopedGoTypeName — for types declared in the current schema
	got := ctx.scopedGoTypeName("ProductOrder")
	if got != "SVC_ProductOrder" {
		t.Errorf("scopedGoTypeName = %q, want SVC_ProductOrder", got)
	}

	// Test currentNsPrefix
	prefix := ctx.currentNsPrefix()
	if prefix != "SVC" {
		t.Errorf("currentNsPrefix = %q, want SVC", prefix)
	}
}

func TestResolveNsScopedGoName_NoScoping(t *testing.T) {
	t.Parallel()

	schema := &xsd.Schema{
		TargetNamespace: "http://example.com/service/v1",
	}

	g := NewGenerator(nil, Config{})
	ctx := newSchemaContext(schema, g)

	got := ctx.scopedGoTypeName("ProductOrder")
	if got != "ProductOrder" {
		t.Errorf("expected plain name without scoping, got %q", got)
	}

	got = ctx.resolveNsScopedGoName("ProductOrder")
	if got != "ProductOrder" {
		t.Errorf("expected plain name without scoping, got %q", got)
	}
}

func TestMapXSDTypeToGoWithContext_NamespaceScoping(t *testing.T) {
	t.Parallel()

	// Create a schema with a complexType and xmlns declarations
	schemaXML := `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:tns="http://example.com/service/v1"
           xmlns:core="http://example.com/core/v1"
           targetNamespace="http://example.com/service/v1">
  <xs:import namespace="http://example.com/core/v1"/>
  <xs:complexType name="MyType">
    <xs:sequence>
      <xs:element name="value" type="xs:string"/>
    </xs:sequence>
  </xs:complexType>
</xs:schema>`

	schema, err := xsd.Parse(strings.NewReader(schemaXML))
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	g := NewGenerator(nil, Config{
		NamespacePrefixes: map[string]string{
			"http://example.com/service/v1": "SVC",
			"http://example.com/core/v1":    "Core",
		},
	})

	ctx := newSchemaContext(schema, g)

	// Local type resolution (tns:MyType or just MyType)
	got := mapXSDTypeToGoWithContext("tns:MyType", ctx)
	if got != "SVC_MyType" {
		t.Errorf("tns:MyType resolved to %q, want SVC_MyType", got)
	}

	// Cross-namespace type resolution (core:SomeType — not in this schema)
	got = mapXSDTypeToGoWithContext("core:SomeType", ctx)
	if got != "Core_SomeType" {
		t.Errorf("core:SomeType resolved to %q, want Core_SomeType", got)
	}

	// XSD builtins should NOT be prefixed
	got = mapXSDTypeToGoWithContext("xs:string", ctx)
	if strings.Contains(got, "_") {
		t.Errorf("xs:string should not be namespace-prefixed, got %q", got)
	}
}
