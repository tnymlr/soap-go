package operation_with_headers

import (
	"encoding/xml"
)

// RequestHeaderWrapper represents the RequestHeader element
type RequestHeaderWrapper struct {
	XMLName   xml.Name `xml:"http://example.com/headers RequestHeader"`
	ContextID string   `xml:"ContextID"`
}

// ResponseHeaderWrapper represents the ResponseHeader element
type ResponseHeaderWrapper struct {
	XMLName       xml.Name `xml:"http://example.com/headers ResponseHeader"`
	TransactionID string   `xml:"TransactionID"`
}

// SubmitOrderWrapper represents the submitOrder element
type SubmitOrderWrapper struct {
	XMLName xml.Name `xml:"http://example.com/headers submitOrder"`
	OrderID string   `xml:"OrderID"`
}

// SubmitOrderResponseWrapper represents the submitOrderResponse element
type SubmitOrderResponseWrapper struct {
	XMLName  xml.Name `xml:"http://example.com/headers submitOrderResponse"`
	Accepted bool     `xml:"Accepted"`
}

// ReserveResourceWrapper represents the reserveResource element
type ReserveResourceWrapper struct {
	XMLName    xml.Name `xml:"http://example.com/headers reserveResource"`
	ResourceID string   `xml:"ResourceID"`
}

// ReserveResourceResponseWrapper represents the reserveResourceResponse element
type ReserveResourceResponseWrapper struct {
	XMLName  xml.Name `xml:"http://example.com/headers reserveResourceResponse"`
	Reserved bool     `xml:"Reserved"`
}
