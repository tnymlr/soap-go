package soapgen

import (
	"strings"

	"github.com/tnymlr/soap-go/internal/codegen"
	"github.com/tnymlr/soap-go/xsd"
)

// generateInlineComplexTypeStruct generates a struct for an inline complex type
func generateInlineComplexTypeStruct(
	g *codegen.File,
	typeName string,
	complexType *xsd.ComplexType,
	ctx *SchemaContext,
) {
	// Add comment
	g.P("// ", typeName, " represents an inline complex type")

	// Start struct declaration
	g.P("type ", typeName, " struct {")

	// Create field registry to track field name collisions
	fieldRegistry := newFieldRegistry()

	hasFields := false

	// Generate fields from the content model (xs:sequence, xs:all, or xs:choice)
	elements, anys := contentModelChildren(complexType)
	if emitContentModelFields(g, elements, anys, ctx, typeName, fieldRegistry) {
		hasFields = true
	}

	// Generate fields from attributes
	for _, attr := range complexType.Attributes {
		if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, typeName) {
			hasFields = true
		}
	}

	// Handle simple content extensions
	if complexType.SimpleContent != nil && complexType.SimpleContent.Extension != nil {
		ext := complexType.SimpleContent.Extension

		// Generate a Value field for the text content based on the extension base
		baseType := mapXSDTypeToGoWithContext(ext.Base, ctx)
		baseType = convertToQualifiedType(baseType, g)
		g.P("\tValue ", baseType, " `xml:\",chardata\"`")
		hasFields = true

		// Handle extension attributes
		for _, attr := range ext.Attributes {
			if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, typeName) {
				hasFields = true
			}
		}
	}

	// If no fields were generated, add a placeholder comment
	if !hasFields {
		g.P("\t// No fields defined")
	}

	// Close struct
	g.P("}")
	g.P()
}

// generateStructFromElement generates a Go struct from an XSD element
func generateStructFromElement(g *codegen.File, element *xsd.Element, ctx *SchemaContext, _ *TypeRegistry) {
	structName := ctx.scopedGoTypeName(element.Name)
	generateStandardStructWithName(g, element, ctx, structName)
}

// generateStructFromElementWithWrapper generates a Go struct from an XSD element with wrapper naming
func generateStructFromElementWithWrapper(
	g *codegen.File,
	element *xsd.Element,
	ctx *SchemaContext,
	_ *TypeRegistry,
) {
	// Generate wrapper-style name
	structName := ctx.scopedGoTypeName(element.Name) + "Wrapper"
	generateStandardStructWithName(g, element, ctx, structName)
}

// generateStandardStructWithName generates a standard struct with a custom struct name
func generateStandardStructWithName(g *codegen.File, element *xsd.Element, ctx *SchemaContext, structName string) {
	// Add comment
	g.P("// ", structName, " represents the ", element.Name, " element")

	// Start struct declaration
	g.P("type ", structName, " struct {")

	// Add XMLName field for namespace handling
	generateXMLNameField(g, element, ctx)

	// Track if we've added any fields
	hasFields := true // XMLName counts as a field

	// Create field registry to track field name collisions
	fieldRegistry := newFieldRegistry()

	// Handle simple type elements (e.g., <element name="foo" type="xsd:string"/>)
	if element.Type != "" && element.ComplexType == nil {
		// Check if this references a complex type
		if complexType := ctx.resolveComplexType(element.Type); complexType != nil {
			// This element references a complex type - embed the complex type's fields directly
			if embedComplexTypeFields(g, complexType, ctx, fieldRegistry, element.Name) {
				hasFields = true
			}
		} else {
			// This is a simple type element, generate a Value field
			goType := mapXSDTypeToGoWithContext(element.Type, ctx)
			goType = convertToQualifiedType(goType, g)
			g.P("\tValue ", goType, " `xml:\",chardata\"`")
			hasFields = true
		}
	} else if element.SimpleType != nil && element.ComplexType == nil {
		// Handle inline simple type elements (e.g., elements with inline enumerations)
		// Check if this inline simple type has been generated as an enum type
		inlineEnumTypeName := ctx.getInlineEnumTypeName(element.Name, element.Name)
		if inlineEnumTypeName != "" {
			// Use the generated inline enum type
			g.P("\tValue ", inlineEnumTypeName, " `xml:\",chardata\"`")
		} else {
			// Fallback to mapping the base type
			baseType := "string" // Default fallback
			if element.SimpleType.Restriction != nil && element.SimpleType.Restriction.Base != "" {
				baseType = mapXSDTypeToGoWithContext(element.SimpleType.Restriction.Base, ctx)
				baseType = convertToQualifiedType(baseType, g)
			}
			g.P("\tValue ", baseType, " `xml:\",chardata\"`")
		}
		hasFields = true
	}

	if element.ComplexType != nil {
		// Handle the top-level content model (xs:sequence, xs:all, or xs:choice)
		elements, anys := contentModelChildren(element.ComplexType)
		if emitContentModelFields(g, elements, anys, ctx, element.Name, fieldRegistry) {
			hasFields = true
		}

		// Handle attributes
		for _, attr := range element.ComplexType.Attributes {
			if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, element.Name) {
				hasFields = true
			}
		}

		// Handle simple content extensions
		if element.ComplexType.SimpleContent != nil && element.ComplexType.SimpleContent.Extension != nil {
			ext := element.ComplexType.SimpleContent.Extension

			// Generate a Value field for the text content based on the extension base
			baseType := mapXSDTypeToGoWithContext(ext.Base, ctx)
			baseType = convertToQualifiedType(baseType, g)
			g.P("\tValue ", baseType, " `xml:\",chardata\"`")
			hasFields = true

			// Handle extension attributes
			for _, attr := range ext.Attributes {
				if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, element.Name) {
					hasFields = true
				}
			}
		}

		// Handle complex content extensions
		if element.ComplexType.ComplexContent != nil && element.ComplexType.ComplexContent.Extension != nil {
			ext := element.ComplexType.ComplexContent.Extension
			extElements, extAnys := extensionContentModelChildren(ext)
			if emitContentModelFields(g, extElements, extAnys, ctx, element.Name, fieldRegistry) {
				hasFields = true
			}

			// Handle extension attributes
			for _, attr := range ext.Attributes {
				if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, element.Name) {
					hasFields = true
				}
			}
		}
	}

	// If no fields were generated beyond XMLName, add a placeholder comment
	if !hasFields {
		g.P("\t// No additional fields defined")
	}

	// Close struct
	g.P("}")
	g.P()
}

// generateStructFromComplexType generates a Go struct from a named complex type
func generateStructFromComplexType(g *codegen.File, complexType *xsd.ComplexType, ctx *SchemaContext) {
	structName := ctx.scopedGoTypeName(complexType.Name)

	// Add comment
	g.P("// ", structName, " represents the ", complexType.Name, " complex type")

	// Start struct declaration
	g.P("type ", structName, " struct {")

	// Create field registry to track field name collisions
	fieldRegistry := newFieldRegistry()

	hasFields := false

	// Handle the top-level content model (xs:sequence, xs:all, or xs:choice)
	elements, anys := contentModelChildren(complexType)
	if emitContentModelFields(g, elements, anys, ctx, complexType.Name, fieldRegistry) {
		hasFields = true
	}

	// Handle attributes
	for _, attr := range complexType.Attributes {
		if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, complexType.Name) {
			hasFields = true
		}
	}

	// Handle simple content extensions
	if complexType.SimpleContent != nil && complexType.SimpleContent.Extension != nil {
		ext := complexType.SimpleContent.Extension

		// Generate a Value field for the text content based on the extension base
		baseType := mapXSDTypeToGoWithContext(ext.Base, ctx)
		baseType = convertToQualifiedType(baseType, g)
		g.P("\tValue ", baseType, " `xml:\",chardata\"`")
		hasFields = true

		// Handle extension attributes
		for _, attr := range ext.Attributes {
			if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, complexType.Name) {
				hasFields = true
			}
		}
	}

	// Handle complex content extensions
	if complexType.ComplexContent != nil && complexType.ComplexContent.Extension != nil {
		ext := complexType.ComplexContent.Extension
		extElements, extAnys := extensionContentModelChildren(ext)
		if emitContentModelFields(g, extElements, extAnys, ctx, complexType.Name, fieldRegistry) {
			hasFields = true
		}

		// Handle extension attributes
		for _, attr := range ext.Attributes {
			if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, complexType.Name) {
				hasFields = true
			}
		}
	}

	// If no fields were generated, add a placeholder comment
	if !hasFields {
		g.P("\t// No fields defined")
	}

	// Close struct
	g.P("}")
	g.P()
}

// generateRawXMLWrapperTypes generates wrapper types for RawXML fields that need their own ,innerxml
func generateRawXMLWrapperTypes(g *codegen.File, ctx *SchemaContext) {
	if ctx == nil || ctx.anonymousTypes == nil {
		return
	}

	generated := make(map[string]bool) // Track generated types to avoid duplicates
	hasTypes := false

	for typeName := range ctx.anonymousTypes {
		// Check if this is a RawXML wrapper type (has RAWXML_ prefix)
		if strings.HasPrefix(typeName, "RAWXML_") && !generated[typeName] {
			// Remove the RAWXML_ prefix to get the actual type name
			actualTypeName := strings.TrimPrefix(typeName, "RAWXML_")

			if !hasTypes {
				g.P("// RawXML wrapper types")
				g.P()
				hasTypes = true
			}

			// Generate a simple wrapper type with a single RawXML field
			g.P("// ", actualTypeName, " represents an inline complex type")
			g.P("type ", actualTypeName, " struct {")
			g.P("\tContent RawXML `xml:\",innerxml\"`")
			g.P("}")
			g.P()
			generated[typeName] = true
		}
	}
}

// embedComplexTypeFields embeds the fields from a complex type into the current struct
func embedComplexTypeFields(
	g *codegen.File,
	complexType *xsd.ComplexType,
	ctx *SchemaContext,
	fieldRegistry *FieldRegistry,
	parentElementName string,
) bool {
	hasFields := false

	// Use the complex type name as parent for anonymous type lookups, falling
	// back to the embedding element name when the complex type is anonymous.
	parentName := complexType.Name
	if parentName == "" {
		parentName = parentElementName
	}

	// Handle the top-level content model (xs:sequence, xs:all, or xs:choice)
	elements, anys := contentModelChildren(complexType)
	if emitContentModelFields(g, elements, anys, ctx, parentName, fieldRegistry) {
		hasFields = true
	}

	// Handle attributes
	for _, attr := range complexType.Attributes {
		if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, parentElementName) {
			hasFields = true
		}
	}

	// Handle simple content extensions
	if complexType.SimpleContent != nil && complexType.SimpleContent.Extension != nil {
		ext := complexType.SimpleContent.Extension

		// Generate a Value field for the text content based on the extension base
		baseType := mapXSDTypeToGoWithContext(ext.Base, ctx)
		baseType = convertToQualifiedType(baseType, g)
		g.P("\tValue ", baseType, " `xml:\",chardata\"`")
		hasFields = true

		// Handle extension attributes
		for _, attr := range ext.Attributes {
			if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, parentElementName) {
				hasFields = true
			}
		}
	}

	// Handle complex content extensions
	if complexType.ComplexContent != nil && complexType.ComplexContent.Extension != nil {
		ext := complexType.ComplexContent.Extension
		extElements, extAnys := extensionContentModelChildren(ext)
		if emitContentModelFields(g, extElements, extAnys, ctx, parentName, fieldRegistry) {
			hasFields = true
		}

		// Handle extension attributes
		for _, attr := range ext.Attributes {
			if generateAttributeFieldWithParentName(g, &attr, ctx, fieldRegistry, parentElementName) {
				hasFields = true
			}
		}
	}

	return hasFields
}
