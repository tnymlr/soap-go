package soap

import (
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// testHeader mimics the shape of a generated SOAP header type: a struct with
// XMLName carrying a namespace, plus simple scalar fields.
type testHeader struct {
	XMLName       xml.Name `xml:"http://example.com/ns MyHeader"`
	ContextID     string   `xml:"ContextID"`
	DealerID      string   `xml:"DealerID"`
	BillingOption string   `xml:"BillingOption"`
}

type testBody struct {
	XMLName xml.Name `xml:"http://example.com/ns MyRequest"`
	Field   string   `xml:"Field"`
}

// testSiblingHeader is a second header entry type used to verify appending.
type testSiblingHeader struct {
	XMLName xml.Name `xml:"http://example.com/other Sibling"`
	Value   string   `xml:"Value"`
}

func TestWithHeaderContent_typedValue(t *testing.T) {
	t.Parallel()
	envelope, err := NewEnvelope(
		WithHeaderContent(&testHeader{
			ContextID:     "ctx-1",
			DealerID:      "TEST-DEALER-1",
			BillingOption: "postpaid",
		}),
		WithBody(&testBody{Field: "value"}),
	)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	got, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	want := strings.Join([]string{
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">`,
		`  <soapenv:Header>`,
		`    <MyHeader xmlns="http://example.com/ns"><ContextID>ctx-1</ContextID><DealerID>TEST-DEALER-1</DealerID><BillingOption>postpaid</BillingOption></MyHeader>`,
		`  </soapenv:Header>`,
		`  <soapenv:Body><MyRequest xmlns="http://example.com/ns"><Field>value</Field></MyRequest></soapenv:Body>`,
		`</soapenv:Envelope>`,
	}, "\n")
	if diff := cmp.Diff(want, strings.TrimSpace(string(got))); diff != "" {
		t.Errorf("envelope XML mismatch (-want +got):\n%s", diff)
	}
}

func TestWithHeaderContent_appendsMultipleEntries(t *testing.T) {
	t.Parallel()
	envelope, err := NewEnvelope(
		WithHeaderContent(&testHeader{ContextID: "c", DealerID: "d", BillingOption: "b"}),
		WithHeaderContent(&testSiblingHeader{Value: "v"}),
	)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if envelope.Header == nil {
		t.Fatal("expected header, got nil")
	}
	if n := len(envelope.Header.Entries); n != 2 {
		t.Fatalf("expected 2 header entries, got %d", n)
	}
	got, err := xml.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	s := string(got)
	if !strings.Contains(s, "<MyHeader xmlns=\"http://example.com/ns\">") {
		t.Errorf("expected MyHeader element in output, got: %s", s)
	}
	if !strings.Contains(s, "<Sibling xmlns=\"http://example.com/other\">") {
		t.Errorf("expected Sibling element in output, got: %s", s)
	}
}

func TestWithHeaderContent_nilIsNoop(t *testing.T) {
	t.Parallel()
	envelope, err := NewEnvelope(
		WithHeaderContent(nil),
		WithBody([]byte(`<x/>`)),
	)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if envelope.Header != nil {
		t.Errorf("expected no Header, got %+v", envelope.Header)
	}
}

func TestWithHeaderContent_acceptsHeaderEntry(t *testing.T) {
	t.Parallel()
	mustUnderstand := true
	entry := HeaderEntry{
		XMLName:        xml.Name{Space: "http://example.com/ns", Local: "Preset"},
		MustUnderstand: &mustUnderstand,
		Actor:          "http://example.com/actor",
		Content:        []byte("<child>x</child>"),
	}
	envelope, err := NewEnvelope(WithHeaderContent(entry))
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if envelope.Header == nil || len(envelope.Header.Entries) != 1 {
		t.Fatalf("expected exactly one header entry, got %+v", envelope.Header)
	}
	got := envelope.Header.Entries[0]
	if got.XMLName != entry.XMLName {
		t.Errorf("XMLName: want %v, got %v", entry.XMLName, got.XMLName)
	}
	if got.MustUnderstand == nil || *got.MustUnderstand != true {
		t.Errorf("MustUnderstand: want true, got %v", got.MustUnderstand)
	}
	if got.Actor != entry.Actor {
		t.Errorf("Actor: want %q, got %q", entry.Actor, got.Actor)
	}
	if string(got.Content) != string(entry.Content) {
		t.Errorf("Content: want %q, got %q", entry.Content, got.Content)
	}
}

func TestWithHeaderContent_nilPointerHeaderEntryErrors(t *testing.T) {
	t.Parallel()
	var entry *HeaderEntry
	_, err := NewEnvelope(WithHeaderContent(entry))
	if err == nil {
		t.Fatal("expected error for nil *HeaderEntry, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("expected error to mention nil, got %q", err.Error())
	}
}

func TestWithHeaderContent_unmarshalableValueErrors(t *testing.T) {
	t.Parallel()
	// Channels cannot be marshalled by encoding/xml.
	ch := make(chan int)
	_, err := NewEnvelope(WithHeaderContent(ch))
	if err == nil {
		t.Fatal("expected error for unmarshalable value, got nil")
	}
	if !strings.Contains(err.Error(), "marshal header content") {
		t.Errorf("expected wrapped marshal error, got %q", err.Error())
	}
}

func TestWithHeaderContent_firstErrorWins(t *testing.T) {
	t.Parallel()
	// Two failing options: the error from the first should survive.
	ch := make(chan int)
	_, err := NewEnvelope(
		WithHeaderContent(ch),
		WithHeaderContent((*HeaderEntry)(nil)),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "marshal header content") {
		t.Errorf("expected first error to win (marshal failure), got %q", err.Error())
	}
}

func TestWithHeader_rawHeaderSetDirectly(t *testing.T) {
	t.Parallel()
	h := &Header{
		Entries: []HeaderEntry{
			{
				XMLName: xml.Name{Space: "http://example.com/ns", Local: "RawHeader"},
				Content: []byte("<x>raw</x>"),
			},
		},
	}
	envelope, err := NewEnvelope(WithHeader(h))
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if envelope.Header == nil {
		t.Fatal("expected header, got nil")
	}
	// The envelope should stamp the Header XMLName with the envelope prefix
	// since the caller didn't specify one.
	if envelope.Header.XMLName.Local != "soapenv:Header" {
		t.Errorf("Header.XMLName.Local: want %q, got %q",
			"soapenv:Header", envelope.Header.XMLName.Local)
	}
}

func TestWithHeader_plusWithHeaderContent_appends(t *testing.T) {
	t.Parallel()
	raw := &Header{
		Entries: []HeaderEntry{
			{
				XMLName: xml.Name{Space: "http://example.com/ns", Local: "First"},
				Content: []byte("<a/>"),
			},
		},
	}
	envelope, err := NewEnvelope(
		WithHeader(raw),
		WithHeaderContent(&testSiblingHeader{Value: "second"}),
	)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if n := len(envelope.Header.Entries); n != 2 {
		t.Fatalf("expected 2 entries, got %d", n)
	}
	if envelope.Header.Entries[0].XMLName.Local != "First" {
		t.Errorf("first entry: want First, got %q", envelope.Header.Entries[0].XMLName.Local)
	}
	if envelope.Header.Entries[1].XMLName.Local != "Sibling" {
		t.Errorf("second entry: want Sibling, got %q", envelope.Header.Entries[1].XMLName.Local)
	}
}

func TestWithHeader_customXMLNamePreserved(t *testing.T) {
	t.Parallel()
	h := &Header{
		XMLName: xml.Name{Local: "SOAP-ENV:Header"},
		Entries: []HeaderEntry{
			{XMLName: xml.Name{Local: "X"}, Content: []byte("")},
		},
	}
	envelope, err := NewEnvelope(WithHeader(h))
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if envelope.Header.XMLName.Local != "SOAP-ENV:Header" {
		t.Errorf("caller's XMLName not preserved: got %q",
			envelope.Header.XMLName.Local)
	}
}

func TestUnmarshalHeaderEntry_roundTripThroughTypedStruct(t *testing.T) {
	t.Parallel()
	// Build an envelope using WithHeaderContent, pull the HeaderEntry back
	// out and decode it into a fresh struct. The two values should match.
	original := &testHeader{
		ContextID:     "ctx-42",
		DealerID:      "TEST-DEALER-A",
		BillingOption: "prepaid",
	}
	env, err := NewEnvelope(WithHeaderContent(original))
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if env.Header == nil || len(env.Header.Entries) != 1 {
		t.Fatalf("expected one header entry, got %+v", env.Header)
	}

	var decoded testHeader
	if err := UnmarshalHeaderEntry(env.Header.Entries[0], &decoded); err != nil {
		t.Fatalf("UnmarshalHeaderEntry: %v", err)
	}
	// encoding/xml populates XMLName on decode from the actual element
	// name; the source struct left it zero-valued (it came from the tag at
	// encode time). Compare only the scalar fields that round-trip through
	// content.
	if decoded.ContextID != original.ContextID ||
		decoded.DealerID != original.DealerID ||
		decoded.BillingOption != original.BillingOption {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, *original)
	}
}

func TestUnmarshalHeaderEntry_preservesNestedContent(t *testing.T) {
	t.Parallel()
	// A header entry built from raw innerxml that holds a nested element
	// with its own namespace declaration must still decode cleanly.
	entry := HeaderEntry{
		XMLName: xml.Name{Space: "http://example.com/ns", Local: "MyHeader"},
		Content: []byte(
			`<ContextID>ctx-1</ContextID><DealerID>TEST-DEALER-B</DealerID><BillingOption>postpaid</BillingOption>`,
		),
	}
	var got testHeader
	if err := UnmarshalHeaderEntry(entry, &got); err != nil {
		t.Fatalf("UnmarshalHeaderEntry: %v", err)
	}
	if got.ContextID != "ctx-1" || got.DealerID != "TEST-DEALER-B" || got.BillingOption != "postpaid" {
		t.Errorf("decoded value mismatch: got %+v", got)
	}
}

func TestUnmarshalHeaderEntry_nonPointerDestErrors(t *testing.T) {
	t.Parallel()
	entry := HeaderEntry{
		XMLName: xml.Name{Space: "http://example.com/ns", Local: "MyHeader"},
	}
	var dest testHeader
	err := UnmarshalHeaderEntry(entry, dest) // pass by value, not pointer
	if err == nil {
		t.Fatal("expected error for non-pointer dest, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal header entry") {
		t.Errorf("expected wrapped unmarshal error, got %q", err.Error())
	}
}

func TestUnmarshalHeaderEntry_mismatchedXMLNameErrors(t *testing.T) {
	t.Parallel()
	// Entry name does not match dest's declared XMLName — encoding/xml
	// surfaces this as an unmarshal error.
	entry := HeaderEntry{
		XMLName: xml.Name{Space: "http://example.com/other", Local: "SomethingElse"},
		Content: []byte("<ContextID>x</ContextID>"),
	}
	var dest testHeader
	err := UnmarshalHeaderEntry(entry, &dest)
	if err == nil {
		t.Fatal("expected error for mismatched XMLName, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal header entry") {
		t.Errorf("expected wrapped unmarshal error, got %q", err.Error())
	}
}

func TestSplitRootElement_selfClosingProducesEmptyContent(t *testing.T) {
	t.Parallel()
	entry, err := splitRootElement([]byte(`<Foo xmlns="http://x"/>`))
	if err != nil {
		t.Fatalf("splitRootElement: %v", err)
	}
	if entry.XMLName.Local != "Foo" || entry.XMLName.Space != "http://x" {
		t.Errorf("XMLName: got %v", entry.XMLName)
	}
	if len(entry.Content) != 0 {
		t.Errorf("expected empty Content for self-closing element, got %q", entry.Content)
	}
}

func TestSplitRootElement_preservesNestedNamespaceDeclarations(t *testing.T) {
	t.Parallel()
	// Verify that nested namespace declarations inside the root element
	// survive the split and end up verbatim in Content.
	data := []byte(`<Outer xmlns="http://o"><Inner xmlns:p="http://p"><p:Leaf/></Inner></Outer>`)
	entry, err := splitRootElement(data)
	if err != nil {
		t.Fatalf("splitRootElement: %v", err)
	}
	inner := string(entry.Content)
	if !strings.Contains(inner, `xmlns:p="http://p"`) {
		t.Errorf("expected inner xmlns:p declaration to be preserved, got %q", inner)
	}
	if !strings.Contains(inner, "<p:Leaf") {
		t.Errorf("expected inner <p:Leaf> to be preserved, got %q", inner)
	}
}

func TestWithHeaderContent_errorIsRetrievable(t *testing.T) {
	t.Parallel()
	// Ensure the surfaced error is wrapped so callers can inspect it.
	_, err := NewEnvelope(WithHeaderContent(make(chan int)))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var xmlErr *xml.UnsupportedTypeError
	if !errors.As(err, &xmlErr) {
		t.Errorf("expected wrapped *xml.UnsupportedTypeError, got %T: %v", err, err)
	}
}
