package soap

import "time"

// XSDDateTime carries an xs:dateTime value while tolerating timezone-less
// inputs, which the XSD spec permits as "unspecified". Timezone-less
// inputs are interpreted as UTC. Marshalling emits RFC 3339 with
// nanosecond precision.
type XSDDateTime struct {
	time.Time
}

func (d *XSDDateTime) UnmarshalText(text []byte) error {
	s := string(text)
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
	}
	var firstErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			d.Time = t
			return nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (d XSDDateTime) MarshalText() ([]byte, error) {
	return []byte(d.Format(time.RFC3339Nano)), nil
}
