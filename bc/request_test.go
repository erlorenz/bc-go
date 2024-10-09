package bc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/erlorenz/bc-go/bc"
	"github.com/erlorenz/bc-go/internal/bctest"
)

func TestMakeRequestGetNoParams(t *testing.T) {

	client, err := bc.NewClient(fakeConfig, bc.WithAuthClient(&fakeTokenGetter{}))
	if err != nil {
		t.Fatalf("failed to create new client: %s", err)
	}

	opts := bc.RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: "fakeEntities",
		QueryParams:   nil,
		Body:          nil,
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
		want := bctest.EmptyString
		if got != want {
			t.Errorf("got %s, wanted %s", got, want)
		}

	})
	// Table of all request pieces
	type testStruct struct {
		name string
		got  any
		want any
	}
	table := []testStruct{
		{"Method", req.Method, "GET"},
		{"Query", req.URL.RawQuery, bctest.EmptyString},
		// Get values and join together with separator so multiple values under same key fail
		{"Header_Accept", strings.Join(req.Header.Values("Accept"), "--"), bc.AcceptJSONNoMetadata},
		{"Header_ContentType", strings.Join(req.Header.Values("Content-Type"), "--"), bctest.EmptyString},
		{"Header_DataAccessIntent", strings.Join(req.Header.Values("Data-Access-Intent"), "--"), bc.DataAccessReadOnly},
		{"Header_IfMatch", strings.Join(req.Header.Values("If-Match"), "--"), bctest.EmptyString},
		// Check entity set name correctly applied
		{"Path", strings.Split(req.URL.Path, "/")[bctest.PathIndexEntitySetName], "fakeEntities"},
	}

	for _, test := range table {
		if test.got != test.want {
			t.Errorf("%s: wanted %v, got %v", test.name, test.want, test.got)
		}
	}

}
func TestMakeRequestGetCommonEndpoint(t *testing.T) {

	commonConfig := fakeConfig
	commonConfig.APIEndpoint = "v2.0"

	client, err := bc.NewClient(commonConfig, bc.WithAuthClient(fakeTokenGetter{}))
	if err != nil {
		t.Fatalf("failed to create new client: %s", err)
	}

	opts := bc.RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: "fakeEntities",
		QueryParams:   nil,
		Body:          nil,
	}

	req, err := client.NewRequest(context.TODO(), opts)
	if err != nil {
		t.Fatal(err)
	}

	// Check entity set name correctly applied
	got := strings.Split(req.URL.Path, "/")[bctest.PathIndexCommonEntitySetName]
	want := "fakeEntities"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

}

func TestMakeRequestPost(t *testing.T) {

	client, err := bc.NewClient(fakeConfig, bc.WithAuthClient(fakeTokenGetter{}))
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

	opts := bc.RequestOptions{
		Method:        http.MethodPost,
		EntitySetName: "fakeEntities",
		Body:          reqBody,
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

	client, err := bc.NewClient(fakeConfig, bc.WithAuthClient(fakeTokenGetter{}))
	if err != nil {
		t.Fatalf("failed to create new client: %s", err)
	}

	qp := bc.QueryParams{
		"param1": "value1",
		"param2": "value2",
	}

	want := "param1=value1&param2=value2"

	opts := bc.RequestOptions{
		Method:        http.MethodGet,
		EntitySetName: "fakeEntibc.EmptyStringties",
		QueryParams:   qp,
		Body:          nil,
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
