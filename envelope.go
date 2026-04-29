package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// Namespace is the standard SOAP 1.1 envelope namespace
const Namespace = "http://schemas.xmlsoap.org/soap/envelope/"

// Envelope represents a SOAP envelope with flexible namespace support.
// It can handle any namespace prefix and URI, making it compatible with various SOAP implementations.
// The XMLName field determines the actual element name and namespace used in marshaling/unmarshaling.
type Envelope struct {
	XMLName xml.Name

	// Optional encoding style as per SOAP 1.1 spec section 4.1.1
	EncodingStyle string `xml:"encodingStyle,attr,omitempty"`

	// Optional header as per SOAP 1.1 spec section 4.2
	Header *Header `xml:"Header,omitempty"`

	// Mandatory body as per SOAP 1.1 spec section 4.3
	Body Body `xml:"Body"`

	// Additional attributes for extensibility as per SOAP 1.1 spec section 4.1
	Attrs []xml.Attr `xml:",any,attr"`
}

// Header represents a SOAP header containing header entries.
// Each header entry can have mustUnderstand and actor attributes as per SOAP 1.1 spec section 4.2.
type Header struct {
	XMLName xml.Name

	// Header entries - flexible content allowing any XML
	Entries []HeaderEntry `xml:",any"`

	// Additional attributes for extensibility
	Attrs []xml.Attr `xml:",any,attr"`
}

// HeaderEntry represents a single header entry with SOAP-specific attributes.
// Implements the mustUnderstand and actor semantics from SOAP 1.1 spec sections 4.2.2 and 4.2.3.
type HeaderEntry struct {
	XMLName xml.Name

	// MustUnderstand attribute as per SOAP 1.1 spec section 4.2.3
	// Values: true (1) means mandatory, false (0) or nil means optional
	MustUnderstand *bool `xml:"mustUnderstand,attr,omitempty"`

	// Actor attribute as per SOAP 1.1 spec section 4.2.2
	// Specifies the intended recipient of this header entry
	Actor string `xml:"actor,attr,omitempty"`

	// Content as raw XML for maximum flexibility
	Content []byte `xml:",innerxml"`

	// Additional attributes for extensibility
	Attrs []xml.Attr `xml:",any,attr"`
}

// Body represents a SOAP body containing the main message payload.
// As per SOAP 1.1 spec section 4.3, it contains body entries.
type Body struct {
	XMLName xml.Name

	// Content as raw XML for maximum flexibility
	// This allows both simple payloads and complex nested structures
	Content []byte `xml:",innerxml"`

	// Additional attributes for extensibility
	Attrs []xml.Attr `xml:",any,attr"`
}

// Fault represents a SOAP fault element as per SOAP 1.1 spec section 4.4.
type Fault struct {
	XMLName xml.Name

	// FaultCode is mandatory and provides algorithmic fault identification
	FaultCode string `xml:"faultcode"`

	// FaultString is mandatory and provides human-readable fault description
	FaultString string `xml:"faultstring"`

	// FaultActor is optional and identifies the fault source
	FaultActor string `xml:"faultactor,omitempty"`

	// Detail is optional and contains application-specific error information
	Detail *Detail `xml:"detail,omitempty"`
}

// String returns a comprehensive string representation of the SOAP fault for logging.
func (f *Fault) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "SOAP fault [%s]: %s", f.FaultCode, f.FaultString)
	if f.FaultActor != "" {
		fmt.Fprintf(&b, " (actor: %s)", f.FaultActor)
	}
	if f.Detail != nil && len(f.Detail.Content) > 0 {
		fmt.Fprintf(&b, " - detail: %s", string(f.Detail.Content))
	}
	return b.String()
}

// Detail represents application-specific fault detail information.
type Detail struct {
	// Content as raw XML to accommodate any application-specific fault data
	Content []byte `xml:",innerxml"`

	// Additional attributes for extensibility
	Attrs []xml.Attr `xml:",any,attr"`
}

type envelopeConfig struct {
	prefix    string
	namespace string
	body      any
	header    *Header
	headerErr error
}

func newEnvelopeConfig() *envelopeConfig {
	return &envelopeConfig{
		prefix:    "soapenv",
		namespace: Namespace,
		body:      nil,
		header:    nil,
	}
}

func (cfg *envelopeConfig) xmlName(element string) xml.Name {
	if cfg.prefix == "" {
		return xml.Name{Local: element}
	}
	return xml.Name{Local: cfg.prefix + ":" + element}
}

func (cfg *envelopeConfig) xmlNSAttr() (xml.Attr, bool) {
	if cfg.prefix == "" || cfg.namespace == "" {
		return xml.Attr{}, false
	}
	return xml.Attr{
		Name:  xml.Name{Local: "xmlns:" + cfg.prefix},
		Value: cfg.namespace,
	}, true
}

// EnvelopeOption is a function that configures an Envelope.
type EnvelopeOption func(*envelopeConfig)

// WithNamespace sets the namespace for the Envelope.
func WithNamespace(prefix, namespace string) EnvelopeOption {
	return func(cfg *envelopeConfig) {
		cfg.prefix = prefix
		cfg.namespace = namespace
	}
}

// WithBody sets the body for the Envelope.
func WithBody(body any) EnvelopeOption {
	return func(cfg *envelopeConfig) {
		cfg.body = body
	}
}

// WithHeader sets a pre-constructed SOAP Header on the Envelope. Use this when
// you need full control over the header (multiple entries, mustUnderstand or
// actor attributes, etc.). For the common case of a single typed header entry,
// prefer WithHeaderContent.
func WithHeader(h *Header) EnvelopeOption {
	return func(cfg *envelopeConfig) {
		cfg.header = h
	}
}

// WithHeaderContent adds a single typed value as a SOAP Header entry. The
// value is marshalled to XML; its root element's name becomes the
// HeaderEntry's XMLName and its children become the HeaderEntry's inner
// content.
//
// Typical usage with generated SOAP types whose XMLName carries the correct
// namespace:
//
//	soap.NewEnvelope(
//	    soap.WithHeaderContent(&MySoapHeader{Field1: "a", Field2: "b"}),
//	    soap.WithBody(&MyRequest{}),
//	)
//
// Multiple WithHeaderContent (and WithHeader) options may be combined to
// construct a header with several entries; entries are appended in the order
// the options are applied.
//
// Limitations: attributes on the value's root element (other than xmlns
// declarations) are not preserved on the HeaderEntry. If you need custom
// attributes, mustUnderstand, or actor, construct a HeaderEntry directly and
// pass it via WithHeader.
//
// If marshalling fails, the error is deferred and returned from NewEnvelope.
// The first such error wins; subsequent header errors are ignored to keep
// diagnostics focused on the original failure.
func WithHeaderContent(v any) EnvelopeOption {
	return func(cfg *envelopeConfig) {
		if cfg.headerErr != nil {
			return
		}
		if v == nil {
			return
		}
		entry, err := headerEntryFromValue(v)
		if err != nil {
			cfg.headerErr = err
			return
		}
		if cfg.header == nil {
			cfg.header = &Header{Entries: []HeaderEntry{entry}}
			return
		}
		cfg.header.Entries = append(cfg.header.Entries, entry)
	}
}

// UnmarshalHeaderEntry decodes a HeaderEntry back into a typed value — the
// response-side counterpart of WithHeaderContent. It marshals the entry
// (preserving the entry's element name, attributes and inner XML) and
// unmarshals the resulting element into dest.
//
// dest must be a non-nil pointer to a struct whose xml:"..." tag on its
// XMLName field matches the entry's element name and namespace; otherwise
// encoding/xml will return an UnmarshalError. Fields on the entry that dest
// does not declare are silently ignored, matching encoding/xml's usual
// relaxed behaviour.
func UnmarshalHeaderEntry(entry HeaderEntry, dest any) error {
	data, err := xml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("soap: marshal header entry: %w", err)
	}
	if err := xml.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("soap: unmarshal header entry: %w", err)
	}
	return nil
}

// headerEntryFromValue converts an arbitrary value into a HeaderEntry. If v
// is itself a HeaderEntry (or *HeaderEntry), it is returned verbatim. Any
// other value is marshalled to XML and its root element is split into name
// and inner content.
func headerEntryFromValue(v any) (HeaderEntry, error) {
	switch x := v.(type) {
	case HeaderEntry:
		return x, nil
	case *HeaderEntry:
		if x == nil {
			return HeaderEntry{}, fmt.Errorf("soap: WithHeaderContent given nil *HeaderEntry")
		}
		return *x, nil
	}
	data, err := xml.Marshal(v)
	if err != nil {
		return HeaderEntry{}, fmt.Errorf("soap: marshal header content: %w", err)
	}
	return splitRootElement(data)
}

// splitRootElement parses a single XML element from data and returns a
// HeaderEntry whose XMLName matches the element's name and whose Content is
// the raw inner XML (children and text, preserving nested namespace
// declarations). Attributes and xmlns declarations on the root element are
// discarded: XMLName.Space is carried forward so the encoder can re-emit the
// namespace, and attribute preservation is an intentionally unsupported case
// (see WithHeaderContent docs).
func splitRootElement(data []byte) (HeaderEntry, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var start xml.StartElement
	var found bool
	for {
		tok, err := dec.Token()
		if err != nil {
			return HeaderEntry{}, fmt.Errorf("soap: read header token: %w", err)
		}
		if s, ok := tok.(xml.StartElement); ok {
			start = s
			found = true
			break
		}
	}
	if !found {
		return HeaderEntry{}, fmt.Errorf("soap: no root element in marshalled header content")
	}
	innerStart := dec.InputOffset()
	depth := 1
	innerEnd := innerStart
	for depth > 0 {
		offsetBefore := dec.InputOffset()
		tok, err := dec.Token()
		if err != nil {
			return HeaderEntry{}, fmt.Errorf("soap: scan header element: %w", err)
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
			if depth == 0 {
				innerEnd = offsetBefore
			}
		}
	}
	var content []byte
	if innerEnd > innerStart {
		content = data[innerStart:innerEnd]
	}
	return HeaderEntry{
		XMLName: start.Name,
		Content: content,
	}, nil
}

// NewEnvelope creates a new SOAP envelope with the specified options.
func NewEnvelope(opts ...EnvelopeOption) (*Envelope, error) {
	cfg := newEnvelopeConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.headerErr != nil {
		return nil, cfg.headerErr
	}
	result := Envelope{
		XMLName: cfg.xmlName("Envelope"),
		Body: Body{
			XMLName: cfg.xmlName("Body"),
		},
	}
	if xmlNSAttr, ok := cfg.xmlNSAttr(); ok {
		result.Attrs = append(result.Attrs, xmlNSAttr)
	}
	switch body := cfg.body.(type) {
	case nil:
		// do nothing
	case []byte:
		result.Body.Content = body
	default:
		bodyData, err := xml.Marshal(body)
		if err != nil {
			return nil, err
		}
		result.Body.Content = bodyData
	}
	if cfg.header != nil {
		hdr := *cfg.header
		// Stamp the Header's XMLName so it inherits the envelope's prefix
		// (e.g. "soapenv:Header"). If the caller supplied one via
		// WithHeader, respect it.
		if hdr.XMLName.Local == "" {
			hdr.XMLName = cfg.xmlName("Header")
		}
		result.Header = &hdr
	}
	return &result, nil
}
