package soapgen

import (
	"regexp"

	"github.com/tnymlr/soap-go/internal/codegen"
	"github.com/tnymlr/soap-go/xsd"
)

// requiresLexicalFormBoolean reports whether a named simpleType restriction
// on xs:boolean has pattern facets that exclude the lexical forms produced
// by Go's stdlib boolean marshaling ("true"/"false"). When this returns
// true the generator must emit a custom Marshal/Unmarshal pair to honor
// the schema's wire-form constraint instead of relying on the default
// bool emit.
func requiresLexicalFormBoolean(simpleType *xsd.SimpleType) bool {
	if simpleType == nil || simpleType.Restriction == nil {
		return false
	}
	base := simpleType.Restriction.Base
	if base != "xs:boolean" && base != "xsd:boolean" {
		return false
	}
	patterns := simpleType.Restriction.Patterns
	if len(patterns) == 0 {
		return false
	}
	return !boolStdlibEmitAcceptedBy(patterns)
}

// boolStdlibEmitAcceptedBy reports whether "true" and "false" are both
// accepted by at least one of the supplied XSD pattern facets (multiple
// xs:pattern facets in one xs:restriction form a union per W3C XSD 1.0
// Datatypes §4.3.4.3).
func boolStdlibEmitAcceptedBy(patterns []xsd.Pattern) bool {
	matchesTrue := false
	matchesFalse := false
	for _, p := range patterns {
		re, err := regexp.Compile(`\A(?:` + p.Value + `)\z`)
		if err != nil {
			continue
		}
		if re.MatchString("true") {
			matchesTrue = true
		}
		if re.MatchString("false") {
			matchesFalse = true
		}
	}
	return matchesTrue && matchesFalse
}

// pickBooleanWireForms returns the literals to emit for true and false
// given the simpleType's xs:pattern facets, plus whether a usable choice
// was found. The numeric forms ("1"/"0") are preferred when both are
// accepted; the alphabetic forms ("true"/"false") are the fallback.
// ok is false when neither pair is fully accepted by the patterns.
func pickBooleanWireForms(patterns []xsd.Pattern) (trueForm, falseForm string, ok bool) {
	matches := func(candidate string) bool {
		for _, p := range patterns {
			re, err := regexp.Compile(`\A(?:` + p.Value + `)\z`)
			if err != nil {
				continue
			}
			if re.MatchString(candidate) {
				return true
			}
		}
		return false
	}
	if matches("1") && matches("0") {
		return "1", "0", true
	}
	if matches("true") && matches("false") {
		return "true", "false", true
	}
	return "", "", false
}

// generateLexicalFormBoolean emits a Go bool-aliased type plus the four
// xml.Marshaler / xml.Unmarshaler methods needed to preserve the wire
// form required by the simpleType's pattern facets. UnmarshalXML and
// UnmarshalXMLAttr accept all four valid xs:boolean lexical forms ("true",
// "false", "1", "0") regardless of which form Marshal produces.
func generateLexicalFormBoolean(g *codegen.File, simpleType *xsd.SimpleType, ctx *SchemaContext) {
	trueForm, falseForm, ok := pickBooleanWireForms(simpleType.Restriction.Patterns)
	if !ok {
		generateRestrictionAlias(g, simpleType, ctx)
		return
	}

	typeName := ctx.scopedGoTypeName(simpleType.Name)
	boolType := g.QualifiedGoIdent(codegen.BoolIdent)
	errorType := g.QualifiedGoIdent(codegen.ErrorIdent)
	stringType := g.QualifiedGoIdent(codegen.StringIdent)
	xmlEncoder := g.QualifiedGoIdent(codegen.XMLEncoderIdent)
	xmlDecoder := g.QualifiedGoIdent(codegen.XMLDecoderIdent)
	xmlStart := g.QualifiedGoIdent(codegen.XMLStartElementIdent)
	xmlAttr := g.QualifiedGoIdent(codegen.XMLAttrIdent)
	xmlName := g.QualifiedGoIdent(codegen.XMLNameIdent)
	fmtErrorf := g.QualifiedGoIdent(codegen.FmtErrorfIdent)

	g.P("// ", typeName, " represents the ", simpleType.Name, " simpleType (restricting xs:boolean)")
	g.P("type ", typeName, " ", boolType)
	g.P()

	g.P("// MarshalXML emits ", typeName, " in its xs:pattern-constrained wire form.")
	g.P("func (b ", typeName, ") MarshalXML(e *", xmlEncoder, ", start ", xmlStart, ") ", errorType, " {")
	g.P("\ts := \"", falseForm, "\"")
	g.P("\tif b {")
	g.P("\t\ts = \"", trueForm, "\"")
	g.P("\t}")
	g.P("\treturn e.EncodeElement(s, start)")
	g.P("}")
	g.P()

	g.P("// UnmarshalXML decodes ", typeName, " from any xs:boolean lexical form.")
	g.P("func (b *", typeName, ") UnmarshalXML(d *", xmlDecoder, ", start ", xmlStart, ") ", errorType, " {")
	g.P("\tvar s ", stringType)
	g.P("\tif err := d.DecodeElement(&s, &start); err != nil {")
	g.P("\t\treturn err")
	g.P("\t}")
	g.P("\tswitch s {")
	g.P("\tcase \"1\", \"true\":")
	g.P("\t\t*b = true")
	g.P("\tcase \"0\", \"false\":")
	g.P("\t\t*b = false")
	g.P("\tdefault:")
	g.P("\t\treturn ", fmtErrorf, "(\"invalid xs:boolean value %q for ", typeName, "\", s)")
	g.P("\t}")
	g.P("\treturn nil")
	g.P("}")
	g.P()

	g.P("// MarshalXMLAttr emits ", typeName, " as an attribute in its xs:pattern-constrained wire form.")
	g.P("func (b ", typeName, ") MarshalXMLAttr(name ", xmlName, ") (", xmlAttr, ", ", errorType, ") {")
	g.P("\ts := \"", falseForm, "\"")
	g.P("\tif b {")
	g.P("\t\ts = \"", trueForm, "\"")
	g.P("\t}")
	g.P("\treturn ", xmlAttr, "{Name: name, Value: s}, nil")
	g.P("}")
	g.P()

	g.P("// UnmarshalXMLAttr decodes ", typeName, " from any xs:boolean attribute lexical form.")
	g.P("func (b *", typeName, ") UnmarshalXMLAttr(attr ", xmlAttr, ") ", errorType, " {")
	g.P("\tswitch attr.Value {")
	g.P("\tcase \"1\", \"true\":")
	g.P("\t\t*b = true")
	g.P("\tcase \"0\", \"false\":")
	g.P("\t\t*b = false")
	g.P("\tdefault:")
	g.P("\t\treturn ", fmtErrorf, "(\"invalid xs:boolean attribute %q for ", typeName, "\", attr.Value)")
	g.P("\t}")
	g.P("\treturn nil")
	g.P("}")
	g.P()
}
