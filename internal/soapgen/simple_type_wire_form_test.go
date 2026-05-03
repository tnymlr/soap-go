package soapgen

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tnymlr/soap-go/xsd"
)

var errInvalidNumericFlag = errors.New("invalid numericFlag value")

type numericFlag bool

func (b numericFlag) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	s := "0"
	if b {
		s = "1"
	}
	return e.EncodeElement(s, start)
}

func (b *numericFlag) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	switch s {
	case "1", "true":
		*b = true
	case "0", "false":
		*b = false
	default:
		return errInvalidNumericFlag
	}
	return nil
}

func (b numericFlag) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	s := "0"
	if b {
		s = "1"
	}
	return xml.Attr{Name: name, Value: s}, nil
}

func (b *numericFlag) UnmarshalXMLAttr(attr xml.Attr) error {
	switch attr.Value {
	case "1", "true":
		*b = true
	case "0", "false":
		*b = false
	default:
		return errInvalidNumericFlag
	}
	return nil
}

type numericFlagWrapper struct {
	XMLName xml.Name    `xml:"Wrapper"`
	Flag    numericFlag `xml:"Flag"`
}

type numericFlagAttrWrapper struct {
	XMLName xml.Name    `xml:"Container"`
	Flag    numericFlag `xml:"flag,attr"`
}

func TestNumericFlag_MarshalEmits01(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   numericFlag
		want string
	}{
		{numericFlag(true), "<Wrapper><Flag>1</Flag></Wrapper>"},
		{numericFlag(false), "<Wrapper><Flag>0</Flag></Wrapper>"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			out, err := xml.Marshal(numericFlagWrapper{Flag: tc.in})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != tc.want {
				t.Errorf("got %q want %q", out, tc.want)
			}
		})
	}
}

func TestNumericFlag_UnmarshalAcceptsAllFourLexicalForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{"<Wrapper><Flag>1</Flag></Wrapper>", true},
		{"<Wrapper><Flag>true</Flag></Wrapper>", true},
		{"<Wrapper><Flag>0</Flag></Wrapper>", false},
		{"<Wrapper><Flag>false</Flag></Wrapper>", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			var w numericFlagWrapper
			if err := xml.Unmarshal([]byte(tc.in), &w); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if bool(w.Flag) != tc.want {
				t.Errorf("got %v want %v", bool(w.Flag), tc.want)
			}
		})
	}
}

func TestNumericFlag_UnmarshalRejectsInvalid(t *testing.T) {
	t.Parallel()
	cases := []string{
		"<Wrapper><Flag>Y</Flag></Wrapper>",
		"<Wrapper><Flag>yes</Flag></Wrapper>",
		"<Wrapper><Flag>2</Flag></Wrapper>",
		"<Wrapper><Flag></Flag></Wrapper>",
		"<Wrapper><Flag>TRUE</Flag></Wrapper>",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			var w numericFlagWrapper
			err := xml.Unmarshal([]byte(in), &w)
			if err == nil {
				t.Errorf("expected error for %q, got nil", in)
			}
		})
	}
}

func TestNumericFlag_MarshalAttrEmits01(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   numericFlag
		want string
	}{
		{numericFlag(true), `<Container flag="1"></Container>`},
		{numericFlag(false), `<Container flag="0"></Container>`},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			out, err := xml.Marshal(numericFlagAttrWrapper{Flag: tc.in})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != tc.want {
				t.Errorf("got %q want %q", out, tc.want)
			}
		})
	}
}

func TestNumericFlag_UnmarshalAttrAcceptsAllFourLexicalForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{`<Container flag="1"/>`, true},
		{`<Container flag="true"/>`, true},
		{`<Container flag="0"/>`, false},
		{`<Container flag="false"/>`, false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			var w numericFlagAttrWrapper
			if err := xml.Unmarshal([]byte(tc.in), &w); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if bool(w.Flag) != tc.want {
				t.Errorf("got %v want %v", bool(w.Flag), tc.want)
			}
		})
	}
}

func TestRequiresLexicalFormBoolean(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *xsd.SimpleType
		want bool
	}{
		{
			name: "nil",
			in:   nil,
			want: false,
		},
		{
			name: "no restriction",
			in:   &xsd.SimpleType{},
			want: false,
		},
		{
			name: "boolean base, no patterns",
			in: &xsd.SimpleType{
				Restriction: &xsd.Restriction{Base: "xs:boolean"},
			},
			want: false,
		},
		{
			name: "string base with pattern",
			in: &xsd.SimpleType{
				Restriction: &xsd.Restriction{
					Base:     "xs:string",
					Patterns: []xsd.Pattern{{Value: "0|1"}},
				},
			},
			want: false,
		},
		{
			name: "boolean base, pattern accepts true and false",
			in: &xsd.SimpleType{
				Restriction: &xsd.Restriction{
					Base:     "xs:boolean",
					Patterns: []xsd.Pattern{{Value: "true|false"}},
				},
			},
			want: false,
		},
		{
			name: "boolean base, pattern accepts only 0 and 1",
			in: &xsd.SimpleType{
				Restriction: &xsd.Restriction{
					Base:     "xs:boolean",
					Patterns: []xsd.Pattern{{Value: "0|1"}},
				},
			},
			want: true,
		},
		{
			name: "xsd:boolean prefix variant",
			in: &xsd.SimpleType{
				Restriction: &xsd.Restriction{
					Base:     "xsd:boolean",
					Patterns: []xsd.Pattern{{Value: "0|1"}},
				},
			},
			want: true,
		},
		{
			name: "two separate patterns covering true and false",
			in: &xsd.SimpleType{
				Restriction: &xsd.Restriction{
					Base: "xs:boolean",
					Patterns: []xsd.Pattern{
						{Value: "true"},
						{Value: "false"},
					},
				},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := requiresLexicalFormBoolean(tc.in)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestPickBooleanWireForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		patterns  []xsd.Pattern
		wantTrue  string
		wantFalse string
		wantOK    bool
	}{
		{
			name:      "0|1",
			patterns:  []xsd.Pattern{{Value: "0|1"}},
			wantTrue:  "1",
			wantFalse: "0",
			wantOK:    true,
		},
		{
			name:      "true|false",
			patterns:  []xsd.Pattern{{Value: "true|false"}},
			wantTrue:  "true",
			wantFalse: "false",
			wantOK:    true,
		},
		{
			name: "split true and false patterns",
			patterns: []xsd.Pattern{
				{Value: "true"},
				{Value: "false"},
			},
			wantTrue:  "true",
			wantFalse: "false",
			wantOK:    true,
		},
		{
			name:     "single constant 0",
			patterns: []xsd.Pattern{{Value: "0"}},
			wantOK:   false,
		},
		{
			name:     "non-canonical Y|N",
			patterns: []xsd.Pattern{{Value: "Y|N"}},
			wantOK:   false,
		},
		{
			name:     "malformed regex",
			patterns: []xsd.Pattern{{Value: "[unclosed"}},
			wantOK:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotTrue, gotFalse, gotOK := pickBooleanWireForms(tc.patterns)
			if gotOK != tc.wantOK {
				t.Errorf("ok: got %v want %v", gotOK, tc.wantOK)
			}
			if gotTrue != tc.wantTrue {
				t.Errorf("trueForm: got %q want %q", gotTrue, tc.wantTrue)
			}
			if gotFalse != tc.wantFalse {
				t.Errorf("falseForm: got %q want %q", gotFalse, tc.wantFalse)
			}
		})
	}
}

func TestGenerateLexicalFormBoolean_GoldenContent(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "boolean_pattern_numeric", "types.go"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	content := string(raw)
	requiredFragments := []string{
		"type NumericFlag bool",
		"func (b NumericFlag) MarshalXML(",
		"func (b *NumericFlag) UnmarshalXML(",
		"func (b NumericFlag) MarshalXMLAttr(",
		"func (b *NumericFlag) UnmarshalXMLAttr(",
		"type VerboseFlag bool",
	}
	for _, frag := range requiredFragments {
		if !strings.Contains(content, frag) {
			t.Errorf("missing required fragment: %q", frag)
		}
	}
	forbiddenForVerbose := []string{
		"func (b VerboseFlag) MarshalXML",
		"func (b *VerboseFlag) UnmarshalXML",
	}
	for _, frag := range forbiddenForVerbose {
		if strings.Contains(content, frag) {
			t.Errorf(
				"VerboseFlag should not have custom marshalers (stdlib bool emit matches its pattern); found %q",
				frag,
			)
		}
	}
}
