package soapgen

import (
	"sort"
	"strings"

	"github.com/tnymlr/soap-go/internal/codegen"
	"github.com/tnymlr/soap-go/xsd"
)

// generateSimpleTypeConstants generates Go type declarations for top-level
// xs:simpleType declarations. Restriction-with-enumerations types become
// typed string enums (with constants, String, IsValid methods). Other
// simpleType shapes — restriction-without-enumerations, xs:list, xs:union —
// are emitted as Go type aliases so that cross-namespace references to
// them compile.
func generateSimpleTypeConstants(g *codegen.File, ctx *SchemaContext) {
	if len(ctx.simpleTypes) == 0 {
		return
	}

	g.P("// Simple types")
	g.P()

	// Sort simple type names for deterministic output
	var names []string
	for name := range ctx.simpleTypes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		simpleType := ctx.simpleTypes[name]
		switch {
		case simpleType.Restriction != nil && len(simpleType.Restriction.Enumerations) > 0:
			generateEnumType(g, simpleType, ctx)
		case simpleType.Restriction != nil:
			generateRestrictionAlias(g, simpleType, ctx)
		case simpleType.List != nil:
			generateListFallback(g, simpleType, ctx)
		case simpleType.Union != nil:
			generateUnionFallback(g, simpleType, ctx)
		default:
			// xs:simpleType with no restriction, list, or union is unusual
			// but emit a string alias so any references compile.
			generateOpaqueFallback(g, simpleType, ctx)
		}
	}
}

// generateRestrictionAlias emits a Go type alias for an xs:simpleType whose
// restriction has no enumeration facets (pattern, maxLength, whiteSpace,
// empty restriction, etc.). The alias targets the Go type of the restriction's
// base XSD type, resolved via mapXSDTypeToGoWithContext so chained
// restrictions collapse to the ultimate base (matching how field-level
// references are mapped).
func generateRestrictionAlias(g *codegen.File, simpleType *xsd.SimpleType, ctx *SchemaContext) {
	typeName := ctx.scopedGoTypeName(simpleType.Name)
	base := simpleType.Restriction.Base
	baseGoType := convertToQualifiedType(mapXSDTypeToGoWithContext(base, ctx), g)

	g.P("// ", typeName, " represents the ", simpleType.Name, " simpleType (restricting ", base, ")")
	g.P("type ", typeName, " ", baseGoType)
	g.P()
}

// generateListFallback emits a placeholder alias for xs:list simpleTypes.
// Proper xs:list support would emit a Go slice type with custom MarshalXML
// and UnmarshalXML methods that split/join on whitespace. No Optus XSDs use
// xs:list, so we defer that work and emit a string alias instead; callers
// must parse the whitespace-separated values manually.
func generateListFallback(g *codegen.File, simpleType *xsd.SimpleType, ctx *SchemaContext) {
	typeName := ctx.scopedGoTypeName(simpleType.Name)
	itemType := simpleType.List.ItemType
	if itemType == "" {
		itemType = "<inline simpleType>"
	}

	g.P("// ", typeName, " represents the ", simpleType.Name, " xs:list simpleType")
	g.P("// (whitespace-separated list of ", itemType, ").")
	g.P("// TODO: implement xs:list with proper slice type + MarshalXML/UnmarshalXML;")
	g.P("// for now emitted as string so cross-namespace references compile.")
	g.P("type ", typeName, " ", g.QualifiedGoIdent(codegen.StringIdent))
	g.P()
}

// generateUnionFallback emits a placeholder alias for xs:union simpleTypes.
// Go has no native sum types and the XML wire form is always a string, so
// even a proper implementation would typically be `type X string` plus
// runtime validation against the member types. No Optus XSDs use xs:union,
// so we defer member-type validation and just emit the string alias.
func generateUnionFallback(g *codegen.File, simpleType *xsd.SimpleType, ctx *SchemaContext) {
	typeName := ctx.scopedGoTypeName(simpleType.Name)

	g.P("// ", typeName, " represents the ", simpleType.Name, " xs:union simpleType")
	if simpleType.Union.MemberTypes != "" {
		g.P("// (union of: ", simpleType.Union.MemberTypes, ").")
	}
	g.P("// TODO: implement xs:union member-type validation;")
	g.P("// for now emitted as string so cross-namespace references compile.")
	g.P("type ", typeName, " ", g.QualifiedGoIdent(codegen.StringIdent))
	g.P()
}

// generateOpaqueFallback handles the rare case of an xs:simpleType that
// carries no restriction, list, or union at all. Emit a string alias so
// references still compile.
func generateOpaqueFallback(g *codegen.File, simpleType *xsd.SimpleType, ctx *SchemaContext) {
	typeName := ctx.scopedGoTypeName(simpleType.Name)

	g.P("// ", typeName, " represents the ", simpleType.Name, " simpleType")
	g.P("type ", typeName, " ", g.QualifiedGoIdent(codegen.StringIdent))
	g.P()
}

// generateEnumType generates a Go enum type from an XSD simple type with enumerations
func generateEnumType(g *codegen.File, simpleType *xsd.SimpleType, ctx *SchemaContext) {
	typeName := ctx.scopedGoTypeName(simpleType.Name)

	// Generate the enum type definition
	g.P("// ", typeName, " represents an enumeration type")
	g.P("type ", typeName, " ", g.QualifiedGoIdent(codegen.StringIdent))
	g.P()

	// Generate the constants with typed values
	g.P("// ", typeName, " enumeration values")
	g.P("const (")

	var enumValues []string
	for _, enum := range simpleType.Restriction.Enumerations {
		constName := typeName + toGoName(enum.Value)
		enumValues = append(enumValues, constName)
		g.P("\t", constName, " ", typeName, " = \"", enum.Value, "\"")
	}
	g.P(")")
	g.P()

	// Generate String method
	g.P("// String returns the string representation of ", typeName)
	g.P("func (e ", typeName, ") String() ", g.QualifiedGoIdent(codegen.StringIdent), " {")
	g.P("\treturn ", g.QualifiedGoIdent(codegen.StringIdent), "(e)")
	g.P("}")
	g.P()

	// Generate IsValid method
	g.P("// IsValid returns true if the ", typeName, " value is valid")
	g.P("func (e ", typeName, ") IsValid() ", g.QualifiedGoIdent(codegen.BoolIdent), " {")
	g.P("\tswitch e {")
	g.P("\tcase ", strings.Join(enumValues, ", "), ":")
	g.P("\t\treturn true")
	g.P("\tdefault:")
	g.P("\t\treturn false")
	g.P("\t}")
	g.P("}")
	g.P()
}

// generateInlineEnumTypes generates Go enum types for inline enumerations
func generateInlineEnumTypes(g *codegen.File, ctx *SchemaContext) {
	if len(ctx.inlineEnums) == 0 {
		return
	}

	g.P("// Inline enumeration types")
	g.P()

	// Collect unique enum types to generate (avoiding duplicates from deduplication)
	uniqueEnums := make(map[string]InlineEnumInfo)
	for _, enumInfo := range ctx.inlineEnums {
		if !enumInfo.Generated {
			uniqueEnums[enumInfo.TypeName] = enumInfo
		}
	}

	// Sort enum type names for deterministic output
	var typeNames []string
	for typeName := range uniqueEnums {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)

	// Generate each unique inline enum type
	for _, typeName := range typeNames {
		enumInfo := uniqueEnums[typeName]
		generateInlineEnumType(g, &enumInfo)

		// Mark all enums with this type name as generated
		for key, info := range ctx.inlineEnums {
			if info.TypeName == typeName {
				info.Generated = true
				ctx.inlineEnums[key] = info
			}
		}
	}
}

// generateInlineEnumType generates a Go enum type from an inline enum info
func generateInlineEnumType(g *codegen.File, enumInfo *InlineEnumInfo) {
	typeName := enumInfo.TypeName
	simpleType := enumInfo.SimpleType

	// Generate the enum type definition
	g.P("// ", typeName, " represents an inline enumeration type")
	g.P("type ", typeName, " ", g.QualifiedGoIdent(codegen.StringIdent))
	g.P()

	// Generate the constants with typed values
	g.P("// ", typeName, " enumeration values")
	g.P("const (")

	var enumValues []string
	for _, enum := range simpleType.Restriction.Enumerations {
		constName := typeName + toGoName(enum.Value)
		enumValues = append(enumValues, constName)
		g.P("\t", constName, " ", typeName, " = \"", enum.Value, "\"")
	}
	g.P(")")
	g.P()

	// Generate String method
	g.P("// String returns the string representation of ", typeName)
	g.P("func (e ", typeName, ") String() ", g.QualifiedGoIdent(codegen.StringIdent), " {")
	g.P("\treturn ", g.QualifiedGoIdent(codegen.StringIdent), "(e)")
	g.P("}")
	g.P()

	// Generate IsValid method
	g.P("// IsValid returns true if the ", typeName, " value is valid")
	g.P("func (e ", typeName, ") IsValid() ", g.QualifiedGoIdent(codegen.BoolIdent), " {")
	g.P("\tswitch e {")
	g.P("\tcase ", strings.Join(enumValues, ", "), ":")
	g.P("\t\treturn true")
	g.P("\tdefault:")
	g.P("\t\treturn false")
	g.P("\t}")
	g.P("}")
	g.P()
}

// generateComplexTypes generates Go structs for named complex types
func generateComplexTypes(g *codegen.File, ctx *SchemaContext) {
	if len(ctx.complexTypes) == 0 {
		return
	}

	g.P("// Complex types")
	g.P()

	// Sort complex type names for deterministic output
	var names []string
	for name := range ctx.complexTypes {
		names = append(names, name)
	}
	sort.Strings(names)

	// Generate each complex type
	for _, name := range names {
		complexType := ctx.complexTypes[name]
		generateStructFromComplexType(g, complexType, ctx)
	}
}
