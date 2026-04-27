package case_insensitive_collisions

import (
	"encoding/xml"
	soap "github.com/justinclift-prvidr/soap-go"
)

// Complex types

// FleetDataType represents the FleetDataType complex type
type FleetDataType struct {
	VehicleCount int32  `xml:"vehicleCount"`
	Status       string `xml:"status"`
}

// RecordType represents the RecordType complex type
type RecordType struct {
	Id        string           `xml:"id"`
	Timestamp soap.XSDDateTime `xml:"timestamp"`
}

// UserRequestType represents the UserRequestType complex type
type UserRequestType struct {
	UserID      string `xml:"userID"`
	RequestType string `xml:"requestType"`
}

// UserRequestWrapper represents the UserRequest element
type UserRequestWrapper struct {
	XMLName     xml.Name `xml:"http://example.com/collisions UserRequest"`
	UserID      string   `xml:"userID"`
	RequestType string   `xml:"requestType"`
}

// GetFleetResponseWrapper represents the GetFleetResponse element
type GetFleetResponseWrapper struct {
	XMLName      xml.Name `xml:"http://example.com/collisions GetFleetResponse"`
	VehicleCount int32    `xml:"vehicleCount"`
	Status       string   `xml:"status"`
}

// DataRecordWrapper represents the DataRecord element
type DataRecordWrapper struct {
	XMLName   xml.Name         `xml:"DataRecord"`
	Id        string           `xml:"id"`
	Timestamp soap.XSDDateTime `xml:"timestamp"`
}
