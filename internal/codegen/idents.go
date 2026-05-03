package codegen

// Common Go identifiers used in generated code.
// These provide type-safe access to commonly used types and functions.
var (
	// Standard library types
	XMLNameIdent         = GoIdent{GoImportPath: "encoding/xml", GoName: "Name"}
	XMLAttrIdent         = GoIdent{GoImportPath: "encoding/xml", GoName: "Attr"}
	XMLEncoderIdent      = GoIdent{GoImportPath: "encoding/xml", GoName: "Encoder"}
	XMLDecoderIdent      = GoIdent{GoImportPath: "encoding/xml", GoName: "Decoder"}
	XMLStartElementIdent = GoIdent{GoImportPath: "encoding/xml", GoName: "StartElement"}
	ContextIdent         = GoIdent{GoImportPath: "context", GoName: "Context"}
	TimeIdent            = GoIdent{GoImportPath: "time", GoName: "Time"}
	BytesBufferIdent     = GoIdent{GoImportPath: "bytes", GoName: "Buffer"}
	HTTPClientIdent      = GoIdent{GoImportPath: "net/http", GoName: "Client"}
	HTTPRequestIdent     = GoIdent{GoImportPath: "net/http", GoName: "Request"}
	HTTPResponseIdent    = GoIdent{GoImportPath: "net/http", GoName: "Response"}
	IOReaderIdent        = GoIdent{GoImportPath: "io", GoName: "Reader"}
	IOReadCloserIdent    = GoIdent{GoImportPath: "io", GoName: "ReadCloser"}

	// Standard library functions
	FmtSprintfIdent                = GoIdent{GoImportPath: "fmt", GoName: "Sprintf"}
	FmtErrorfIdent                 = GoIdent{GoImportPath: "fmt", GoName: "Errorf"}
	XMLMarshalIdent                = GoIdent{GoImportPath: "encoding/xml", GoName: "Marshal"}
	XMLUnmarshalIdent              = GoIdent{GoImportPath: "encoding/xml", GoName: "Unmarshal"}
	HTTPNewRequestWithContextIdent = GoIdent{GoImportPath: "net/http", GoName: "NewRequestWithContext"}
	HTTPStatusOKIdent              = GoIdent{GoImportPath: "net/http", GoName: "StatusOK"}
	BytesNewReaderIdent            = GoIdent{GoImportPath: "bytes", GoName: "NewReader"}
	IOReadAllIdent                 = GoIdent{GoImportPath: "io", GoName: "ReadAll"}

	// SOAP library types
	SOAPClientIdent               = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "Client"}
	SOAPClientOptionIdent         = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "ClientOption"}
	SOAPNewClientIdent            = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "NewClient"}
	SOAPWithEndpointIdent         = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "WithEndpoint"}
	SOAPEnvelopeIdent             = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "Envelope"}
	SOAPBodyIdent                 = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "Body"}
	SOAPNamespaceIdent            = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "Namespace"}
	SOAPNewEnvelopeIdent          = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "NewEnvelope"}
	SOAPWithBodyIdent             = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "WithBody"}
	SOAPWithHeaderContentIdent    = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "WithHeaderContent"}
	SOAPUnmarshalHeaderEntryIdent = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "UnmarshalHeaderEntry"}
	SOAPXSDDateTimeIdent          = GoIdent{GoImportPath: "github.com/tnymlr/soap-go", GoName: "XSDDateTime"}

	// Built-in types (no import path needed)
	StringIdent = GoIdent{GoImportPath: "", GoName: "string"}
	IntIdent    = GoIdent{GoImportPath: "", GoName: "int"}
	BoolIdent   = GoIdent{GoImportPath: "", GoName: "bool"}
	ByteIdent   = GoIdent{GoImportPath: "", GoName: "byte"}
	ErrorIdent  = GoIdent{GoImportPath: "", GoName: "error"}
)
