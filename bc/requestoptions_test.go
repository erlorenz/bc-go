package bc_test

import (
	"net/http"
	"testing"

	"github.com/erlorenz/bc-go/bc"
)

func TestRequestOptions(t *testing.T) {
	type testCase struct {
		name string
		opts bc.RequestOptions
		want bool //should pass
	}
	tests := []testCase{
		{
			name: "valid get",
			want: true,
			opts: bc.RequestOptions{
				Method:        http.MethodGet,
				EntitySetName: "fakeEntities"},
		},
		{
			name: "valid get with params",
			want: true,
			opts: bc.RequestOptions{
				Method: http.MethodGet,
				QueryParams: bc.QueryParams{
					"$filter": "number eq 'XXXX'",
				},
			},
		},
		{
			name: "valid post no body",
			want: true,
			opts: bc.RequestOptions{
				Method:        http.MethodPost,
				EntitySetName: "fakeEntities"},
		},
		{
			name: "valid post with body",
			want: true,
			opts: bc.RequestOptions{
				Method: http.MethodPost,
				Body:   "a non-nil body"},
		},
	}

	for _, test := range tests {
		err := test.opts.Validate()
		// Pass if err is nil
		got := err == nil
		if test.want != got {
			t.Logf("%s - %s", test.name, err)
			t.Errorf("unexpected result - want %t, got %t", test.want, got)
		}

	}

}
