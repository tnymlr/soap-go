package soapgen

import (
	"strings"

	"github.com/tnymlr/soap-go/internal/codegen"
	"github.com/tnymlr/soap-go/xsd"
)

// buildXMLTag constructs an XML struct tag with appropriate omitempty behavior
func buildXMLTag(xmlName string, isOptional bool, isAttribute bool) string {
	// Trim spaces from XML name to avoid suspicious struct tag warnings
	xmlName = strings.TrimSpace(xmlName)
	parts := []string{xmlName}

	if isAttribute {
		parts = append(parts, "attr")
	}

	if isOptional {
		parts = append(parts, "omitempty")
	}

	return strings.Join(parts, ",")
}

// generateXMLNameField generates an XMLName field with appropriate namespace handling
func generateXMLNameField(g *codegen.File, element *xsd.Element, ctx *SchemaContext) {
	// Trim spaces from element name to avoid suspicious struct tag warnings
	elementName := strings.TrimSpace(element.Name)

	// For operation elements (used in SOAP messages), include the target namespace
	// This ensures proper WSDL compliance for both requests and responses
	if ctx.generator != nil && ctx.generator.isOperationMessageElement(elementName) {
		if ctx.schema.TargetNamespace != "" {
			g.P(
				"\tXMLName ",
				g.QualifiedGoIdent(codegen.XMLNameIdent),
				" `xml:\"",
				ctx.schema.TargetNamespace,
				" ",
				elementName,
				"\"`",
			)
		} else {
			g.P("\tXMLName ", g.QualifiedGoIdent(codegen.XMLNameIdent), " `xml:\"", elementName, "\"`")
		}
	} else {
		// For non-operation elements, use flexible namespace handling
		g.P("\tXMLName ", g.QualifiedGoIdent(codegen.XMLNameIdent), " `xml:\"", elementName, "\"`")
	}
}

// convertToQualifiedType converts raw type strings to use QualifiedGoIdent for proper import management
func convertToQualifiedType(rawType string, g *codegen.File) string {
	switch rawType {
	case "time.Time":
		return g.QualifiedGoIdent(codegen.TimeIdent)
	case "soap.XSDDateTime":
		return g.QualifiedGoIdent(codegen.SOAPXSDDateTimeIdent)
	case "string":
		return g.QualifiedGoIdent(codegen.StringIdent)
	case "bool":
		return g.QualifiedGoIdent(codegen.BoolIdent)
	case "int":
		return g.QualifiedGoIdent(codegen.IntIdent)
	case "[]byte":
		return "[]" + g.QualifiedGoIdent(codegen.ByteIdent)
	default:
		return rawType
	}
}

// contentModelChildren returns the elements and xs:any entries from a
// complex type's top-level content model — xs:sequence, xs:all, or
// xs:choice. At most one of these is non-nil per the XSD spec.
//
// xs:all in XSD 1.0 cannot contain xs:any, so the anys slice is always
// nil for All. xs:choice can contain both elements and xs:any siblings.
func contentModelChildren(complexType *xsd.ComplexType) (elements []xsd.Element, anys []xsd.Any) {
	if complexType == nil {
		return nil, nil
	}
	if complexType.Sequence != nil {
		return complexType.Sequence.Elements, complexType.Sequence.Any
	}
	if complexType.All != nil {
		return complexType.All.Elements, nil
	}
	if complexType.Choice != nil {
		return complexType.Choice.Elements, complexType.Choice.Any
	}
	return nil, nil
}

// extensionContentModelChildren is the equivalent of contentModelChildren
// for an xs:extension inside xs:complexContent.
func extensionContentModelChildren(ext *xsd.Extension) (elements []xsd.Element, anys []xsd.Any) {
	if ext == nil {
		return nil, nil
	}
	if ext.Sequence != nil {
		return ext.Sequence.Elements, ext.Sequence.Any
	}
	if ext.All != nil {
		return ext.All.Elements, nil
	}
	if ext.Choice != nil {
		return ext.Choice.Elements, ext.Choice.Any
	}
	return nil, nil
}

// shouldUseRawXMLForComplexType determines if a complex type should be represented as RawXML
// instead of generating a structured type. This is true for complex types that contain xs:any
// elements anywhere in their content model, or whose children are all bare-untyped.
func shouldUseRawXMLForComplexType(complexType *xsd.ComplexType) bool {
	elements, anys := contentModelChildren(complexType)

	// xs:any anywhere in the content model means we cannot generate a typed struct.
	if len(anys) > 0 {
		return true
	}

	if len(elements) == 0 {
		return false
	}

	// Check if all elements are untyped (no type attribute, no inline complex type,
	// and no ref to a globally-declared element).
	hasTypedElements := false
	for _, elem := range elements {
		if elem.Type != "" || elem.ComplexType != nil || elem.Ref != "" {
			hasTypedElements = true
			break
		}
	}

	// If there are only untyped elements, use RawXML.
	return !hasTypedElements
}
