package bc

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateOf(t *testing.T) {

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatalf("coudnt load timezone 'America/Chicago': %s", err)
	}
	// Use 18:30 so it is next day UTC
	dt := time.Date(2024, time.February, 18, 18, 30, 0, 0, loc)
	d := DateOf(dt)

	if d.Year != 2024 || d.Month != time.February || d.Day != 18 {
		t.Errorf("wanted '2024-02-18', got '%s'", d)
	}

	// Change to UTC and check that it is the next date
	dtUTC := dt.In(time.UTC)
	dUTC := DateOf(dtUTC)

	if dUTC.Year != 2024 || dUTC.Month != time.February || dUTC.Day != 19 {
		t.Errorf("wanted '2024-02-19', got '%s'", d)
	}

}

func TestParseGoodDate(t *testing.T) {

	goodDate := "2024-02-18"
	_, err := ParseDate(goodDate)
	if err != nil {
		t.Error(err)
	}

}

func TestParseBadDate(t *testing.T) {

	badDate := "2024---02-18"
	_, err := ParseDate(badDate)
	if err == nil {
		t.Fatalf("no error parsing '%s', wanted err", badDate)
	}
}

func TestTimeUTC(t *testing.T) {
	date, err := ParseDate("2024-02-01")
	if err != nil {
		t.Fatal(err)
	}

	time := date.TimeUTC()

	if time.Month() != 2 {
		t.Errorf("wrong month, expected 2, got %d", time.Month())
	}
	if time.Day() != 1 {
		t.Errorf("wrong date, expected 1, got %d", time.Day())
	}
}

func TestMarshalFromParseDate(t *testing.T) {

	d, err := ParseDate("2024-02-18")
	if err != nil {
		t.Fatal(err)
	}

	b, err := d.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	want := `"2024-02-18"`
	got := string(b)

	if got != want {
		t.Errorf("wanted %s, got %s", want, got)
	}
}

func TestMarshalFromDateOf(t *testing.T) {

	dt := time.Date(2024, time.February, 20, 0, 0, 0, 0, time.UTC)
	d := DateOf(dt)

	b, err := d.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	want := `"2024-02-20"`
	got := string(b)

	if got != want {
		t.Errorf("wanted %s, got %s", want, got)
	}
}

func TestUnmarshal(t *testing.T) {

	bytes := []byte(`{"date":"2024-02-18"}`)

	var dStruct struct {
		Date Date `json:"date"`
	}

	err := json.Unmarshal(bytes, &dStruct)
	if err != nil {
		t.Fatal(err)
	}

	want := "2024-02-18"
	got := dStruct.Date.String()

	if got != want {
		t.Errorf("wanted %s, got %s", want, got)
	}
}
