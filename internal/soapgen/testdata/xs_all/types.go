package xs_all

import (
	"encoding/xml"
)

// AllInline_Nested represents an inline complex type
type AllInline_Nested struct {
	Inner1 string `xml:"Inner1"`
	Inner2 string `xml:"Inner2"`
}

// Inline complex types

// AllSeqAll_Outer represents an inline complex type
type AllSeqAll_Outer struct {
	Middle AllSeqAllOuter_Middle `xml:"Middle"`
}

// AllSeqAllOuter_Middle represents an inline complex type
type AllSeqAllOuter_Middle struct {
	LeafA string `xml:"LeafA"`
	LeafB *int32 `xml:"LeafB,omitempty"`
}

// Complex types

// AllNamed represents the AllNamed complex type
type AllNamed struct {
	LabelA string         `xml:"LabelA"`
	LabelB *string        `xml:"LabelB,omitempty"`
	Detail *DetailWrapper `xml:"Detail,omitempty"`
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

// AllInlineWrapper represents the AllInline element
type AllInlineWrapper struct {
	XMLName  xml.Name          `xml:"AllInline"`
	Status   *StatusWrapper    `xml:"Status,omitempty"`
	FieldA   string            `xml:"FieldA"`
	FieldB   int32             `xml:"FieldB"`
	Nested   *AllInline_Nested `xml:"Nested,omitempty"`
	Revision *string           `xml:"revision,attr,omitempty"`
}

// AllNamedRefWrapper represents the AllNamedRef element
type AllNamedRefWrapper struct {
	XMLName xml.Name       `xml:"AllNamedRef"`
	LabelA  string         `xml:"LabelA"`
	LabelB  *string        `xml:"LabelB,omitempty"`
	Detail  *DetailWrapper `xml:"Detail,omitempty"`
}

// AllSeqAllWrapper represents the AllSeqAll element
type AllSeqAllWrapper struct {
	XMLName xml.Name        `xml:"AllSeqAll"`
	Outer   AllSeqAll_Outer `xml:"Outer"`
}
