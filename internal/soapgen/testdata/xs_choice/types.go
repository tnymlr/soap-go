package xs_choice

import (
	"encoding/xml"
)

// ChoiceInline_OptionD represents an inline complex type
type ChoiceInline_OptionD struct {
	Sub1 string `xml:"Sub1"`
	Sub2 int32  `xml:"Sub2"`
}

// Inline complex types

// Complex types

// ChoiceNamed represents the ChoiceNamed complex type
type ChoiceNamed struct {
	First  string `xml:"First"`
	Second int32  `xml:"Second"`
}

// OptionAWrapper represents the OptionA element
type OptionAWrapper struct {
	XMLName xml.Name `xml:"OptionA"`
	Value   string   `xml:",chardata"`
}

// OptionBWrapper represents the OptionB element
type OptionBWrapper struct {
	XMLName xml.Name `xml:"OptionB"`
	Value   int32    `xml:",chardata"`
}

// ChoiceInlineWrapper represents the ChoiceInline element
type ChoiceInlineWrapper struct {
	XMLName xml.Name             `xml:"ChoiceInline"`
	OptionA OptionAWrapper       `xml:"OptionA"`
	OptionB OptionBWrapper       `xml:"OptionB"`
	OptionC string               `xml:"OptionC"`
	OptionD ChoiceInline_OptionD `xml:"OptionD"`
}

// ChoiceNamedRefWrapper represents the ChoiceNamedRef element
type ChoiceNamedRefWrapper struct {
	XMLName xml.Name `xml:"ChoiceNamedRef"`
	First   string   `xml:"First"`
	Second  int32    `xml:"Second"`
}

// ChoiceWithAnyWrapper represents the ChoiceWithAny element
type ChoiceWithAnyWrapper struct {
	XMLName xml.Name `xml:"ChoiceWithAny"`
	Known   string   `xml:"Known"`
	Content RawXML   `xml:",innerxml"`
}
