package bc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMakeRequestGetNoParams(t *testing.T) {

	client, err := NewClient(fakeConfig, fakeToken)
	if err != nil {
		t.Fatalf("failed to create new client: %s", err)
	}

	opts := MakeRequestOptions{
		method:        http.MethodGet,
		entitySetName: "fakeEntities",
		queryParams:   nil,
		body:          nil,
	}

	req, err := client.NewRequest(context.TODO(), opts)
	if err != nil {
		t.Fatal(err)
	}

	// Test body separate from table
	t.Run("Body", func(t *testing.T) {
		buf := &bytes.Buffer{}
		if _, err := io.ReadAll(buf); err != nil {
			t.Fatalf("failed reading body into buffer")
		}
		got := buf.String()
		want := EmptyString
		if got != want {
			t.Errorf("got %s, wanted %s", got, want)
		}

	})
	// Table of all request pieces
	table := []TestNameGotWant[string]{
		{"Method", req.Method, "GET"},
		{"Query", req.URL.RawQuery, EmptyString},
		// Get values and join together with separator so multiple values under same key fail
		{"Header_Accept", strings.Join(req.Header.Values("Accept"), "--"), AcceptJSONNoMetadata},
		{"Header_ContentType", strings.Join(req.Header.Values("Content-Type"), "--"), EmptyString},
		{"Header_DataAccessIntent", strings.Join(req.Header.Values("Data-Access-Intent"), "--"), DataAccessReadOnly},
		{"Header_IfMatch", strings.Join(req.Header.Values("If-Match"), "--"), EmptyString},
		// Check entity set name correctly applied
		{"Path", strings.Split(req.URL.Path, "/")[pathIndexEntitySetName], "fakeEntities"},
	}

	RunTable(table, t)

}
func TestMakeRequestGetCommonEndpoint(t *testing.T) {

	commonConfig := fakeConfig
	commonConfig.APIEndpoint = "v2.0"

	client, err := NewClient(commonConfig, fakeToken)
	if err != nil {
		t.Fatalf("failed to create new client: %s", err)
	}

	opts := MakeRequestOptions{
		method:        http.MethodGet,
		entitySetName: "fakeEntities",
		queryParams:   nil,
		body:          nil,
	}

	req, err := client.NewRequest(context.TODO(), opts)
	if err != nil {
		t.Fatal(err)
	}

	// Table of all request pieces
	table := []TestNameGotWant[string]{
		// Check entity set name correctly applied
		{"Path", strings.Split(req.URL.Path, "/")[pathIndexCommonEntitySetName], "fakeEntities"},
	}

	RunTable(table, t)

}

func TestMakeRequestPost(t *testing.T) {

	client, err := NewClient(fakeConfig, fakeToken)
	if err != nil {
		t.Fatalf("failed to create new client: %s", err)
	}

	type DTO struct {
		Name    string
		Age     int
		Blocked bool
	}

	reqBody := DTO{
		Name:    "Fred",
		Age:     30,
		Blocked: false,
	}

	opts := MakeRequestOptions{
		method:        http.MethodGet,
		entitySetName: "fakeEntities",
		body:          reqBody,
	}

	req, err := client.NewRequest(context.TODO(), opts)
	if err != nil {
		t.Fatal(err)
	}

	got := DTO{}
	if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode json: %s", err)
	}

	want := reqBody
	if got != want {
		t.Errorf("wanted %+v, got %+v", want, got)
	}

}

func TestMakeRequestGetParams(t *testing.T) {

	client, err := NewClient(fakeConfig, fakeToken)
	if err != nil {
		t.Fatalf("failed to create new client: %s", err)
	}

	qp := QueryParams{
		"param1": "value1",
		"param2": "value2",
	}

	want := "param1=value1&param2=value2"

	opts := MakeRequestOptions{
		method:        http.MethodGet,
		entitySetName: "fakeEntities",
		queryParams:   qp,
		body:          nil,
	}

	req, err := client.NewRequest(context.TODO(), opts)
	if err != nil {
		t.Fatal(err)
	}

	got := req.URL.RawQuery

	if got != want {
		t.Errorf("wanted %s, got %s", want, got)
	}

}
