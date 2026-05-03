package boolean_pattern_numeric

import (
	"encoding/xml"
	"fmt"
)

// Simple types

// NumericFlag represents the NumericFlag simpleType (restricting xs:boolean)
type NumericFlag bool

// MarshalXML emits NumericFlag in its xs:pattern-constrained wire form.
func (b NumericFlag) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	s := "0"
	if b {
		s = "1"
	}
	return e.EncodeElement(s, start)
}

// UnmarshalXML decodes NumericFlag from any xs:boolean lexical form.
func (b *NumericFlag) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	switch s {
	case "1", "true":
		*b = true
	case "0", "false":
		*b = false
	default:
		return fmt.Errorf("invalid xs:boolean value %q for NumericFlag", s)
	}
	return nil
}

// MarshalXMLAttr emits NumericFlag as an attribute in its xs:pattern-constrained wire form.
func (b NumericFlag) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	s := "0"
	if b {
		s = "1"
	}
	return xml.Attr{Name: name, Value: s}, nil
}

// UnmarshalXMLAttr decodes NumericFlag from any xs:boolean attribute lexical form.
func (b *NumericFlag) UnmarshalXMLAttr(attr xml.Attr) error {
	switch attr.Value {
	case "1", "true":
		*b = true
	case "0", "false":
		*b = false
	default:
		return fmt.Errorf("invalid xs:boolean attribute %q for NumericFlag", attr.Value)
	}
	return nil
}

// VerboseFlag represents the VerboseFlag simpleType (restricting xsd:boolean)
type VerboseFlag bool
