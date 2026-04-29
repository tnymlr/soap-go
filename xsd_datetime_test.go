package soap

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestXSDDateTime_UnmarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "timezone-less hundredths fractional",
			input: "2026-04-16T17:16:17.00",
			want:  time.Date(2026, 4, 16, 17, 16, 17, 0, time.UTC),
		},
		{
			name:  "timezone-less no fractional",
			input: "2026-04-16T17:16:17",
			want:  time.Date(2026, 4, 16, 17, 16, 17, 0, time.UTC),
		},
		{
			name:  "timezone-less nanosecond fractional",
			input: "2026-04-16T17:16:17.123456789",
			want:  time.Date(2026, 4, 16, 17, 16, 17, 123456789, time.UTC),
		},
		{
			name:  "RFC 3339 with Z",
			input: "2026-04-16T17:16:17Z",
			want:  time.Date(2026, 4, 16, 17, 16, 17, 0, time.UTC),
		},
		{
			name:  "RFC 3339 with offset",
			input: "2026-04-16T17:16:17+10:00",
			want:  time.Date(2026, 4, 16, 17, 16, 17, 0, time.FixedZone("", 10*3600)),
		},
		{
			name:  "RFC 3339 nano with Z",
			input: "2026-04-16T17:16:17.123456789Z",
			want:  time.Date(2026, 4, 16, 17, 16, 17, 123456789, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var d XSDDateTime
			if err := d.UnmarshalText([]byte(tt.input)); err != nil {
				t.Fatalf("UnmarshalText(%q) error: %v", tt.input, err)
			}
			if !d.Equal(tt.want) {
				t.Errorf("UnmarshalText(%q) = %v, want %v", tt.input, d.Time, tt.want)
			}
		})
	}
}

func TestXSDDateTime_UnmarshalText_Error(t *testing.T) {
	t.Parallel()

	var d XSDDateTime
	if err := d.UnmarshalText([]byte("not a timestamp")); err == nil {
		t.Errorf("expected error for invalid input, got nil")
	}
}

func TestXSDDateTime_MarshalText(t *testing.T) {
	t.Parallel()

	d := XSDDateTime{Time: time.Date(2026, 4, 16, 17, 16, 17, 0, time.UTC)}
	got, err := d.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	want := "2026-04-16T17:16:17Z"
	if string(got) != want {
		t.Errorf("MarshalText = %q, want %q", string(got), want)
	}
}

// TestXSDDateTime_XMLRoundTrip exercises the type through encoding/xml,
// which is how generated code uses it. Both element text and attributes
// route through TextMarshaler / TextUnmarshaler.
func TestXSDDateTime_XMLRoundTrip(t *testing.T) {
	t.Parallel()

	type record struct {
		XMLName xml.Name    `xml:"record"`
		When    XSDDateTime `xml:"when,attr"`
		Body    XSDDateTime `xml:"body"`
	}

	input := []byte(`<record when="2026-04-16T17:16:17.00"><body>2026-04-16T23:59:59.00</body></record>`)
	var r record
	if err := xml.Unmarshal(input, &r); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	wantAttr := time.Date(2026, 4, 16, 17, 16, 17, 0, time.UTC)
	if !r.When.Equal(wantAttr) {
		t.Errorf("attr = %v, want %v", r.When.Time, wantAttr)
	}

	wantBody := time.Date(2026, 4, 16, 23, 59, 59, 0, time.UTC)
	if !r.Body.Equal(wantBody) {
		t.Errorf("body = %v, want %v", r.Body.Time, wantBody)
	}

	out, err := xml.Marshal(r)
	if err != nil {
		t.Fatalf("xml.Marshal: %v", err)
	}
	wantOut := `<record when="2026-04-16T17:16:17Z"><body>2026-04-16T23:59:59Z</body></record>`
	if string(out) != wantOut {
		t.Errorf("Marshal = %q, want %q", string(out), wantOut)
	}
}
