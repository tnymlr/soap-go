package soapgen

import (
	"fmt"
	"strings"

	"github.com/tnymlr/soap-go/internal/codegen"
	"github.com/tnymlr/soap-go/wsdl"
)

// generateClientFile generates a Go file with SOAP client implementation
func (g *Generator) generateClientFile(packageName, filename string) (*codegen.File, error) {
	// Check if there are any SOAP operations to generate client for
	soapBindings := g.getSOAPBindings()
	if len(soapBindings) == 0 {
		// No SOAP bindings found, don't generate client file
		return nil, nil
	}

	// Check if any bindings have operations
	hasOperations := false
	for _, binding := range soapBindings {
		portType := g.getPortTypeForBinding(binding)
		if portType != nil && len(portType.Operations) > 0 {
			hasOperations = true
			break
		}
	}

	if !hasOperations {
		// No operations found, don't generate client file
		return nil, nil
	}

	file := codegen.NewFile(filename, packageName)

	// Set custom package name for soap-go to use "soap" instead of "soapgo"
	file.SetPackageName("github.com/tnymlr/soap-go", "soap")

	// Add package declaration
	file.P("package ", packageName)
	file.P()

	// Imports are now handled automatically via QualifiedGoIdent calls

	// Generate client option types
	g.generateClientOptions(file)

	// Generate client struct
	g.generateClientStruct(file)

	// Generate NewClient function
	g.generateNewClientFunction(file)

	// Generate operation methods
	err := g.generateOperationMethods(file)
	if err != nil {
		return nil, fmt.Errorf("failed to generate operation methods: %w", err)
	}

	// Generate helper functions
	g.generateHelperFunctions(file)

	return file, nil
}

// generateClientOptions generates type aliases for SOAP client options
func (g *Generator) generateClientOptions(file *codegen.File) {
	file.P("// ClientOption configures a Client.")
	file.P("type ClientOption = ", file.QualifiedGoIdent(codegen.SOAPClientOptionIdent))
	file.P()
}

// generateClientStruct generates the main Client struct
func (g *Generator) generateClientStruct(file *codegen.File) {
	file.P("// Client is a SOAP client for this service.")
	file.P("type Client struct {")
	file.P("\t*", file.QualifiedGoIdent(codegen.SOAPClientIdent))
	file.P("}")
	file.P()
}

// generateNewClientFunction generates the NewClient constructor
func (g *Generator) generateNewClientFunction(file *codegen.File) {
	// Extract default endpoint from service definitions
	endpoint := g.getDefaultEndpoint()

	file.P("// NewClient creates a new SOAP client.")
	file.P("func NewClient(opts ...ClientOption) (*Client, error) {")
	if endpoint != "" {
		file.P("\tsoapOpts := append([]", file.QualifiedGoIdent(codegen.SOAPClientOptionIdent), "{")
		file.P("\t\t", file.QualifiedGoIdent(codegen.SOAPWithEndpointIdent), "(\"", endpoint, "\"),")
		file.P("\t}, opts...)")
		file.P("\tsoapClient, err := ", file.QualifiedGoIdent(codegen.SOAPNewClientIdent), "(soapOpts...)")
	} else {
		file.P("\tsoapClient, err := ", file.QualifiedGoIdent(codegen.SOAPNewClientIdent), "(opts...)")
	}
	file.P("\tif err != nil {")
	file.P(
		"\t\treturn nil, ",
		file.QualifiedGoIdent(codegen.FmtErrorfIdent),
		"(\"failed to create SOAP client: %w\", err)",
	)
	file.P("\t}")
	file.P("\treturn &Client{")
	file.P("\t\tClient: soapClient,")
	file.P("\t}, nil")
	file.P("}")
	file.P()
}

// getDefaultEndpoint extracts the default endpoint from service definitions
func (g *Generator) getDefaultEndpoint() string {
	for _, service := range g.definitions.Service {
		for _, port := range service.Ports {
			if port.SOAP11Address != nil {
				return port.SOAP11Address.Location
			}
			if port.SOAP12Address != nil {
				return port.SOAP12Address.Location
			}
		}
	}
	return ""
}

// generateOperationMethods generates methods for each SOAP operation
func (g *Generator) generateOperationMethods(file *codegen.File) error {
	// Find SOAP bindings and their operations
	soapBindings := g.getSOAPBindings()

	for _, binding := range soapBindings {
		portType := g.getPortTypeForBinding(binding)
		if portType == nil {
			continue
		}

		for _, operation := range portType.Operations {
			err := g.generateOperationMethod(file, &operation, binding)
			if err != nil {
				return fmt.Errorf("failed to generate method for operation %s: %w", operation.Name, err)
			}
		}
	}

	return nil
}

// getSOAPBindings returns SOAP bindings from the WSDL, preferring SOAP 1.1 over SOAP 1.2
func (g *Generator) getSOAPBindings() []*wsdl.Binding {
	var soap11Bindings []*wsdl.Binding
	var soap12Bindings []*wsdl.Binding

	for i := range g.definitions.Binding {
		binding := &g.definitions.Binding[i]
		if binding.SOAP11Binding != nil {
			soap11Bindings = append(soap11Bindings, binding)
		} else if binding.SOAP12Binding != nil {
			soap12Bindings = append(soap12Bindings, binding)
		}
	}

	// Prefer SOAP 1.1 bindings as per README.md specification
	if len(soap11Bindings) > 0 {
		return soap11Bindings
	}

	// Fall back to SOAP 1.2 if no SOAP 1.1 bindings are available
	return soap12Bindings
}

// getPortTypeForBinding finds the port type that matches the given binding
func (g *Generator) getPortTypeForBinding(binding *wsdl.Binding) *wsdl.PortType {
	// Extract the local name from the binding type (remove namespace prefix if present)
	bindingType := binding.Type
	if colonIdx := strings.LastIndex(bindingType, ":"); colonIdx != -1 {
		bindingType = bindingType[colonIdx+1:]
	}

	for i := range g.definitions.PortType {
		portType := &g.definitions.PortType[i]
		if portType.Name == bindingType {
			return portType
		}
	}
	return nil
}

// generateOperationMethod generates a single operation method
func (g *Generator) generateOperationMethod(
	file *codegen.File,
	operation *wsdl.Operation,
	binding *wsdl.Binding,
) error {
	methodName := toGoName(operation.Name)

	// Check if this is a one-way operation (no output message)
	isOneWay := operation.Output == nil

	// Get input and output message types
	inputType, outputType, err := g.getOperationTypes(operation)
	if err != nil {
		return fmt.Errorf("failed to get types for operation %s: %w", operation.Name, err)
	}

	// Get SOAP action
	soapAction := g.getSOAPActionForOperation(operation.Name, binding)

	// Look up the matching binding operation so we can inspect its
	// <soap:header> declarations.
	var bindingOp *wsdl.BindingOperation
	for i := range binding.BindingOperations {
		if binding.BindingOperations[i].Name == operation.Name {
			bindingOp = &binding.BindingOperations[i]
			break
		}
	}

	// Resolve input/output header types. SOAP 1.2 headers are not yet
	// surfaced to the generator since no operation in scope uses them.
	var inputHeader *headerType
	if bindingOp != nil && bindingOp.Input != nil && len(bindingOp.Input.SOAP11Header) > 0 {
		inputHeader, err = g.resolveHeaderType(&bindingOp.Input.SOAP11Header[0])
		if err != nil {
			return fmt.Errorf("failed to resolve input header for operation %s: %w", operation.Name, err)
		}
	}
	var outputHeader *headerType
	if !isOneWay && bindingOp != nil && bindingOp.Output != nil && len(bindingOp.Output.SOAP11Header) > 0 {
		outputHeader, err = g.resolveHeaderType(&bindingOp.Output.SOAP11Header[0])
		if err != nil {
			return fmt.Errorf("failed to resolve output header for operation %s: %w", operation.Name, err)
		}
	}

	// When the operation declares an output header, bundle the typed
	// header and body together in a generated <OpName>Result struct.
	resultTypeName := ""
	if outputHeader != nil {
		resultTypeName = methodName + "Result"
		file.P("// ", resultTypeName, " bundles the typed SOAP Header and Body returned by ", methodName, ".")
		file.P("type ", resultTypeName, " struct {")
		file.P("\tHeader *", outputHeader.GoName)
		file.P("\tBody   *", outputType)
		file.P("}")
		file.P()
	}

	// Generate method signature and documentation
	if operation.Documentation != "" {
		// Clean up documentation
		doc := strings.TrimSpace(operation.Documentation)
		doc = strings.ReplaceAll(doc, "\n", " ")
		file.P("// ", methodName, " ", doc)
	} else {
		if isOneWay {
			file.P("// ", methodName, " executes the ", operation.Name, " one-way SOAP operation.")
		} else {
			file.P("// ", methodName, " executes the ", operation.Name, " SOAP operation.")
		}
	}

	// Emit the method signature. Shape depends on three axes: one-way vs
	// request-response, whether an input header is declared, and whether
	// an output header is declared (implying a <OpName>Result return type).
	sig := "func (c *Client) " + methodName + "(ctx " + file.QualifiedGoIdent(codegen.ContextIdent)
	if inputHeader != nil {
		sig += ", header *" + inputHeader.GoName
	}
	sig += ", req *" + inputType + ", opts ...ClientOption)"
	switch {
	case isOneWay:
		sig += " " + file.QualifiedGoIdent(codegen.ErrorIdent) + " {"
	case outputHeader != nil:
		sig += " (*" + resultTypeName + ", " + file.QualifiedGoIdent(codegen.ErrorIdent) + ") {"
	default:
		sig += " (*" + outputType + ", " + file.QualifiedGoIdent(codegen.ErrorIdent) + ") {"
	}
	file.P(sig)

	// Envelope construction — prepend WithHeaderContent when an input
	// header is declared.
	if inputHeader != nil {
		file.P(
			"\treqEnvelope, err := ",
			file.QualifiedGoIdent(codegen.SOAPNewEnvelopeIdent),
			"(",
			file.QualifiedGoIdent(codegen.SOAPWithHeaderContentIdent),
			"(header), ",
			file.QualifiedGoIdent(codegen.SOAPWithBodyIdent),
			"(req))",
		)
	} else {
		file.P(
			"\treqEnvelope, err := ",
			file.QualifiedGoIdent(codegen.SOAPNewEnvelopeIdent),
			"(",
			file.QualifiedGoIdent(codegen.SOAPWithBodyIdent),
			"(req))",
		)
	}
	if isOneWay {
		file.P("\tif err != nil {")
		file.P(
			"\t\treturn ",
			file.QualifiedGoIdent(codegen.FmtErrorfIdent),
			"(\"failed to create SOAP envelope: %w\", err)",
		)
		file.P("\t}")
	} else {
		file.P("\tif err != nil {")
		file.P(
			"\t\treturn nil, ",
			file.QualifiedGoIdent(codegen.FmtErrorfIdent),
			"(\"failed to create SOAP envelope: %w\", err)",
		)
		file.P("\t}")
	}

	// Call + response handling.
	callAction := "\"\""
	if soapAction != "" {
		callAction = "\"" + soapAction + "\""
	}
	if isOneWay {
		file.P("\t_, err = c.Call(ctx, ", callAction, ", reqEnvelope, opts...)")
		file.P("\tif err != nil {")
		file.P("\t\treturn ", file.QualifiedGoIdent(codegen.FmtErrorfIdent), "(\"SOAP call failed: %w\", err)")
		file.P("\t}")
		file.P("\treturn nil")
	} else {
		file.P("\trespEnvelope, err := c.Call(ctx, ", callAction, ", reqEnvelope, opts...)")
		file.P("\tif err != nil {")
		file.P("\t\treturn nil, ", file.QualifiedGoIdent(codegen.FmtErrorfIdent), "(\"SOAP call failed: %w\", err)")
		file.P("\t}")
		file.P("\tvar result ", outputType)
		file.P(
			"\tif err := ",
			file.QualifiedGoIdent(codegen.XMLUnmarshalIdent),
			"(respEnvelope.Body.Content, &result); err != nil {",
		)
		file.P(
			"\t\treturn nil, ",
			file.QualifiedGoIdent(codegen.FmtErrorfIdent),
			"(\"failed to unmarshal response body: %w\", err)",
		)
		file.P("\t}")

		if outputHeader == nil {
			file.P("\treturn &result, nil")
		} else {
			// Locate and decode the declared response header, then
			// bundle it into the generated Result wrapper. Strict:
			// a missing header is an error, per project decision.
			file.P("\tif respEnvelope.Header == nil {")
			file.P(
				"\t\treturn nil, ",
				file.QualifiedGoIdent(codegen.FmtErrorfIdent),
				"(\"expected response header {%s}%s, envelope has no header\", \"",
				outputHeader.Namespace,
				"\", \"",
				outputHeader.LocalName,
				"\")",
			)
			file.P("\t}")
			file.P("\tvar respHeader ", outputHeader.GoName)
			file.P("\tfoundHeader := false")
			file.P("\tfor _, entry := range respEnvelope.Header.Entries {")
			file.P(
				"\t\tif entry.XMLName.Space == \"",
				outputHeader.Namespace,
				"\" && entry.XMLName.Local == \"",
				outputHeader.LocalName,
				"\" {",
			)
			file.P(
				"\t\t\tif err := ",
				file.QualifiedGoIdent(codegen.SOAPUnmarshalHeaderEntryIdent),
				"(entry, &respHeader); err != nil {",
			)
			file.P(
				"\t\t\t\treturn nil, ",
				file.QualifiedGoIdent(codegen.FmtErrorfIdent),
				"(\"failed to unmarshal response header: %w\", err)",
			)
			file.P("\t\t\t}")
			file.P("\t\t\tfoundHeader = true")
			file.P("\t\t\tbreak")
			file.P("\t\t}")
			file.P("\t}")
			file.P("\tif !foundHeader {")
			file.P(
				"\t\treturn nil, ",
				file.QualifiedGoIdent(codegen.FmtErrorfIdent),
				"(\"expected response header {%s}%s, not found\", \"",
				outputHeader.Namespace,
				"\", \"",
				outputHeader.LocalName,
				"\")",
			)
			file.P("\t}")
			file.P("\treturn &", resultTypeName, "{Header: &respHeader, Body: &result}, nil")
		}
	}
	file.P("}")
	file.P()

	return nil
}

// getOperationTypes determines the input and output types for an operation
func (g *Generator) getOperationTypes(operation *wsdl.Operation) (inputType, outputType string, err error) {
	// Get input type
	if operation.Input != nil {
		inputType, err = g.getMessageElementType(operation.Input.Message)
		if err != nil {
			return "", "", fmt.Errorf("failed to get input type: %w", err)
		}
	}

	// Get output type
	if operation.Output != nil {
		outputType, err = g.getMessageElementType(operation.Output.Message)
		if err != nil {
			return "", "", fmt.Errorf("failed to get output type: %w", err)
		}
	}

	// Provide default types if not found
	if inputType == "" {
		inputType = "interface{}"
	}
	if outputType == "" {
		// For operations without output messages, use an empty struct
		// This is more appropriate than interface{} for acknowledgment responses
		outputType = "struct{}"
	}

	return inputType, outputType, nil
}

// getMessageElementType gets the Go type name for a message element
func (g *Generator) getMessageElementType(messageName string) (string, error) {
	// Remove namespace prefix if present
	if colonIdx := strings.LastIndex(messageName, ":"); colonIdx != -1 {
		messageName = messageName[colonIdx+1:]
	}

	// Get the binding style for consistent naming
	bindingStyle := g.getBindingStyle()

	// Find the message definition
	for _, message := range g.definitions.Messages {
		if message.Name == messageName {
			// Get the element from the message part
			if len(message.Parts) > 0 {
				part := message.Parts[0]
				if part.Element != "" {
					// Resolve the QName reference to its namespace URI.
					// In cross-namespace WSDLs the message part's element
					// may live in a different namespace than the WSDL's
					// targetNamespace, so we resolve the prefix via the
					// WSDL's xmlns declarations rather than assuming.
					nsURI, elementName := g.definitions.ResolveQName(part.Element)
					return g.getConsistentTypeName(elementName, nsURI, bindingStyle), nil
				}
			}
		}
	}

	return "", fmt.Errorf("message %s not found", messageName)
}

// headerType describes a WSDL <soap:header> reference resolved to its Go type
// plus the XML namespace/local name the element carries. The namespace and
// local name are needed at response-decode time to match the incoming
// HeaderEntry against the declared header element.
type headerType struct {
	GoName    string
	Namespace string
	LocalName string
}

// resolveHeaderType resolves a <soap:header>'s message+part reference to a
// Go type name and the corresponding XML element identity. Unlike body
// resolution (which implicitly takes the first part of the message), header
// resolution honours the part name declared on the binding — WSDL permits
// multi-part messages and <soap:header> must name which part carries the
// header payload.
func (g *Generator) resolveHeaderType(h *wsdl.SOAPHeader) (*headerType, error) {
	messageName := h.Message
	if colonIdx := strings.LastIndex(messageName, ":"); colonIdx != -1 {
		messageName = messageName[colonIdx+1:]
	}

	bindingStyle := g.getBindingStyle()

	for _, message := range g.definitions.Messages {
		if message.Name != messageName {
			continue
		}
		for _, part := range message.Parts {
			if part.Name != h.Part {
				continue
			}
			if part.Element == "" {
				return nil, fmt.Errorf(
					"header part %q of message %q has no element reference",
					h.Part, messageName,
				)
			}
			nsURI, elementName := g.definitions.ResolveQName(part.Element)
			return &headerType{
				GoName:    g.getConsistentTypeName(elementName, nsURI, bindingStyle),
				Namespace: nsURI,
				LocalName: elementName,
			}, nil
		}
		return nil, fmt.Errorf("header part %q not found in message %q", h.Part, messageName)
	}
	return nil, fmt.Errorf("header message %q not found", messageName)
}

// getSOAPActionForOperation gets the SOAP action for an operation from binding
func (g *Generator) getSOAPActionForOperation(operationName string, binding *wsdl.Binding) string {
	for _, bindingOp := range binding.BindingOperations {
		if bindingOp.Name == operationName {
			if bindingOp.SOAP11Operation != nil {
				return bindingOp.SOAP11Operation.SOAPAction
			}
			if bindingOp.SOAP12Operation != nil {
				return bindingOp.SOAP12Operation.SOAPAction
			}
		}
	}
	return ""
}

// generateHelperFunctions generates helper types and functions for SOAP
func (g *Generator) generateHelperFunctions(file *codegen.File) {
	// Note: SOAP envelope types are now provided by the public API
	// No need to generate private types anymore
}
