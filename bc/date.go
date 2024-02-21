package bc

import (
	"encoding/json"
	"fmt"
	"time"
)

// Date represents a Date type in Business Central.
// It has no time zone associated with so does not represent a unique moment.
// When converting a time.Time to this Date make sure that it is set with the correct time.Location.
// It can be marshaled and unmarshaled and satisfies the Stringer interface.
// Heavily inspired by the civil package:
// https://github.com/googleapis/google-cloud-go/blob/v0.112.0/civil/civil.go
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// DateOf transforms a time.Time into a Date using that
// time.Time's location.
func DateOf(time time.Time) Date {
	var d Date
	d.Year, d.Month, d.Day = time.Date()
	return d
}

// ParseDate transforms a string format 'YYYY-MM-DD' to a Date.
func ParseDate(s string) (Date, error) {

	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("failed to parse time: %w", err)
	}
	return DateOf(t), nil
}

// String formats it 'YYYY-MM-DD'
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// UnmarshalJSON takes the date string (formatted 'YYYY-MM-DD') and converts it to a Date.
func (d *Date) UnmarshalJSON(data []byte) error {
	// Unmarshal as string
	var v string
	err := json.Unmarshal(data, &v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal into string: %w", err)
	}

	date, err := ParseDate(v)
	if err != nil {
		return err
	}
	*d = date
	return nil

}

// MarshalJSON just returns it as a string formatted 'YYYY-MM-DD'.
func (d Date) MarshalJSON() ([]byte, error) {

	// Marshal as string
	b, err := json.Marshal(d.String())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Date: %w", err)
	}

	return b, nil
}

// IsZero returns true if the Date is set to the zero value.
func (d Date) IsZero() bool {
	return (d.Year == 0) && (int(d.Month) == 0) && (d.Day == 0)
}
