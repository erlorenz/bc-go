package bc

import (
	"fmt"
	"strings"
	"time"
)

// Date represents a Date type in Business Central.
// It has no time zone associated with so does not represent a unique moment.
// When converting a time.Time to this Date make sure that it is set with the correct time.Location.
// It can be marshalled and unmarshalled and meets the Stringer interface.
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
	// Trim out the quotes if it is json
	s = strings.Trim(s, `"`)

	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, err
	}
	return DateOf(t), nil
}

// String formats it 'YYYY-MM-DD'
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// UnmarshalJSON takes the date string (formatted 'YYYY-MM-DD') and converts it to a Date.
func (d *Date) UnmarshalJSON(data []byte) error {
	date, err := ParseDate(string(data))
	if err != nil {
		return err
	}
	*d = date
	return nil

}

// MarshalJSON just returns it as a string formatted 'YYYY-MM-DD'.
func (d Date) MarshalJSON() ([]byte, error) {

	// Add the quotes
	str := fmt.Sprintf(`"%s"`, d.String())

	return []byte(str), nil
}

// IsZero returns true if the Date is set to the zero value.
func (d Date) IsZero() bool {
	return (d.Year == 0) && (int(d.Month) == 0) && (d.Day == 0)
}
