package inline_complex_type_with_refs

import (
	"encoding/xml"
)

// RawXML captures raw XML content for untyped elements.
type RawXML []byte

// OnlyRefs_Inner represents an inline complex type
type OnlyRefs_Inner struct {
	Status StatusWrapper  `xml:"Status"`
	Detail *DetailWrapper `xml:"Detail,omitempty"`
}

// Inline complex types

// MixedUntypedAndRefs_Inner represents an inline complex type
type MixedUntypedAndRefs_Inner struct {
	Name   string        `xml:"Name"`
	Status StatusWrapper `xml:"Status"`
}

// StatusWrapper represents the Status element
type StatusWrapper struct {
	XMLName xml.Name `xml:"Status"`
	Value   string   `xml:",chardata"`
}

// DetailWrapper represents the Detail element
type DetailWrapper struct {
	XMLName xml.Name `xml:"Detail"`
	Value   string   `xml:",chardata"`
}

// OnlyRefsWrapper represents the OnlyRefs element
type OnlyRefsWrapper struct {
	XMLName xml.Name       `xml:"OnlyRefs"`
	Inner   OnlyRefs_Inner `xml:"Inner"`
}

// MixedUntypedAndRefsWrapper represents the MixedUntypedAndRefs element
type MixedUntypedAndRefsWrapper struct {
	XMLName xml.Name                  `xml:"MixedUntypedAndRefs"`
	Inner   MixedUntypedAndRefs_Inner `xml:"Inner"`
}
