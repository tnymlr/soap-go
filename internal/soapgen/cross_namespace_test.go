package soapgen

import (
	"encoding/xml"
	"testing"

	"github.com/tnymlr/soap-go/wsdl"
)

// TestGetMessageElementType_CrossNamespace verifies that when a WSDL message
// part references an element in a different namespace than the WSDL's
// targetNamespace (via an xmlns prefix on <definitions>), the client-side
// generated Go type name uses the prefix for the element's namespace, not
// the WSDL's namespace. This must match what the struct generator does for
// the type declaration; a mismatch produces uncompilable output.
func TestGetMessageElementType_CrossNamespace(t *testing.T) {
	t.Parallel()

	defs := &wsdl.Definitions{
		TargetNamespace: "http://example.com/operations",
		ExtraAttrs: []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "ts"}, Value: "http://example.com/data"},
		},
		Messages: []wsdl.Message{
			{
				Name: "addPolicy_RequestMessage",
				Parts: []wsdl.Part{
					{Name: "parameter", Element: "ts:addPolicyRequest"},
				},
			},
		},
	}

	g := NewGenerator(defs, Config{
		NamespacePrefixes: map[string]string{
			"http://example.com/operations": "",
			"http://example.com/data":       "DATA",
		},
	})

	got, err := g.getMessageElementType("addPolicy_RequestMessage")
	if err != nil {
		t.Fatalf("getMessageElementType: %v", err)
	}
	want := "DATA_AddPolicyRequestWrapper"
	if got != want {
		t.Errorf("got %q, want %q — client must use element's own namespace, not WSDL's targetNamespace", got, want)
	}
}

// TestGetMessageElementType_UnprefixedReference verifies that an unprefixed
// element reference falls back to the WSDL's targetNamespace for prefix
// resolution.
func TestGetMessageElementType_UnprefixedReference(t *testing.T) {
	t.Parallel()

	defs := &wsdl.Definitions{
		TargetNamespace: "http://example.com/operations",
		Messages: []wsdl.Message{
			{
				Name: "localMessage",
				Parts: []wsdl.Part{
					{Name: "parameter", Element: "localElement"},
				},
			},
		},
	}

	g := NewGenerator(defs, Config{
		NamespacePrefixes: map[string]string{
			"http://example.com/operations": "OPS",
		},
	})

	got, err := g.getMessageElementType("localMessage")
	if err != nil {
		t.Fatalf("getMessageElementType: %v", err)
	}
	want := "OPS_LocalElementWrapper"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
