package inline_types_in_named_complex_types

import (
	"encoding/xml"
	soap "github.com/justinclift-prvidr/soap-go"
)

// RawXML captures raw XML content for untyped elements.
type RawXML []byte

// ResponseType_Data represents an inline complex type
type ResponseType_Data struct {
	Id       string                    `xml:"id"`
	Value    string                    `xml:"value"`
	Metadata ResponseTypeData_Metadata `xml:"metadata"`
}

// ResponseTypeData_Metadata represents an inline complex type
type ResponseTypeData_Metadata struct {
	Timestamp soap.XSDDateTime `xml:"timestamp"`
	Source    string           `xml:"source"`
}

// ResponseType_Items represents an inline complex type
type ResponseType_Items struct {
	ItemId   string `xml:"itemId"`
	ItemName string `xml:"itemName"`
}

// Inline complex types

// Complex types

// ResponseType represents the ResponseType complex type
type ResponseType struct {
	Status string               `xml:"status"`
	Data   ResponseType_Data    `xml:"data"`
	Items  []ResponseType_Items `xml:"items"`
}

// ResponseWrapper represents the Response element
type ResponseWrapper struct {
	XMLName xml.Name             `xml:"Response"`
	Status  string               `xml:"status"`
	Data    ResponseType_Data    `xml:"data"`
	Items   []ResponseType_Items `xml:"items"`
}
