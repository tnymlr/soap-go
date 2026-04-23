package wsdl_test

import (
	"encoding/xml"
	"testing"

	"github.com/tnymlr/soap-go/wsdl"
)

func TestDefinitions_NamespacePrefixMap(t *testing.T) {
	t.Parallel()
	defs := &wsdl.Definitions{
		ExtraAttrs: []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "ts"}, Value: "http://example.com/data"},
			{Name: xml.Name{Space: "xmlns", Local: "header"}, Value: "http://example.com/hdr"},
			{Name: xml.Name{Space: "", Local: "name"}, Value: "not-an-xmlns"},
		},
	}

	got := defs.NamespacePrefixMap()
	want := map[string]string{
		"ts":     "http://example.com/data",
		"header": "http://example.com/hdr",
	}

	if len(got) != len(want) {
		t.Fatalf("prefix count mismatch: got %d, want %d (%v)", len(got), len(want), got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("prefix %q: got %q, want %q", k, got[k], v)
		}
	}
}

func TestDefinitions_ResolveQName(t *testing.T) {
	t.Parallel()
	defs := &wsdl.Definitions{
		TargetNamespace: "http://example.com/operations",
		ExtraAttrs: []xml.Attr{
			{Name: xml.Name{Space: "xmlns", Local: "ts"}, Value: "http://example.com/data"},
			{Name: xml.Name{Space: "xmlns", Local: "header"}, Value: "http://example.com/hdr"},
		},
	}

	tests := []struct {
		name      string
		ref       string
		wantNSURI string
		wantLocal string
	}{
		{
			name:      "prefixed reference resolves to declared namespace",
			ref:       "ts:addPolicyRequest",
			wantNSURI: "http://example.com/data",
			wantLocal: "addPolicyRequest",
		},
		{
			name:      "different prefix resolves correctly",
			ref:       "header:SoapHeaderMsg",
			wantNSURI: "http://example.com/hdr",
			wantLocal: "SoapHeaderMsg",
		},
		{
			name:      "unprefixed reference falls back to WSDL targetNamespace",
			ref:       "someLocalElement",
			wantNSURI: "http://example.com/operations",
			wantLocal: "someLocalElement",
		},
		{
			name:      "unknown prefix returns empty namespace",
			ref:       "unknown:Element",
			wantNSURI: "",
			wantLocal: "Element",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNS, gotLocal := defs.ResolveQName(tc.ref)
			if gotNS != tc.wantNSURI {
				t.Errorf("nsURI: got %q, want %q", gotNS, tc.wantNSURI)
			}
			if gotLocal != tc.wantLocal {
				t.Errorf("localName: got %q, want %q", gotLocal, tc.wantLocal)
			}
		})
	}
}
