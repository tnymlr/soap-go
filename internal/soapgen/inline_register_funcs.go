package soapgen

import (
	"github.com/tnymlr/soap-go/xsd"
)

// registerInlineTypesFromElement recursively registers inline complex types from an element without generating struct definitions
func registerInlineTypesFromElement(
	element *xsd.Element,
	_ string,
	ctx *SchemaContext,
	registry *AnonymousTypeRegistry,
) {
	if element.ComplexType == nil {
		return
	}
	elements, _ := contentModelChildren(element.ComplexType)
	registerInlineTypesFromElements(elements, element.Name, ctx, registry)
}

// registerInlineTypesFromComplexType recursively registers inline complex types from a complex type without generating struct definitions
func registerInlineTypesFromComplexType(
	complexType *xsd.ComplexType,
	parentName string,
	ctx *SchemaContext,
	registry *AnonymousTypeRegistry,
) {
	elements, _ := contentModelChildren(complexType)
	registerInlineTypesFromElements(elements, parentName, ctx, registry)
}

// registerInlineTypesFromElements walks content-model children and
// registers names for each inline complex type found, recursing into
// their content models as well.
func registerInlineTypesFromElements(
	elements []xsd.Element,
	parentName string,
	ctx *SchemaContext,
	registry *AnonymousTypeRegistry,
) {
	for _, field := range elements {
		if field.ComplexType == nil {
			continue
		}
		// Generate type name using the same logic as the generation pass
		// Don't use the registry to avoid conflicts - just compute the name directly
		typeName := toGoName(parentName) + "_" + toGoName(field.Name)
		ctx.anonymousTypes[typeName] = true

		// Recursively register nested inline types
		registerInlineTypesFromComplexType(field.ComplexType, typeName, ctx, registry)
	}
}
