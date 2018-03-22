package sender_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/helper"
	"github.com/ubuntu/ubuntu-report/internal/sender"
)

func TestGetURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		baseURL string

		want    string
		wantErr bool
	}{
		{"regular", "https://myurl.com", "https://myurl.com/distroname/desktop/versionnumber", false},
		{"bad parsing", "http://a b.com/", "", true},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			got, err := sender.GetURL(tc.baseURL, "distroname", "versionnumber")

			a.CheckWantedErr(err, tc.wantErr)
			if err != nil {
				return
			}
			a.Equal(got, tc.want)
		})
	}
}

func TestSend(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		status  int
		wantErr bool
	}{
		{http.StatusOK, false},
		{http.StatusNotFound, true},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(string(tc.status), func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			status := statusHandler(tc.status)
			ts := httptest.NewServer(&status)
			defer ts.Close()

			err := sender.Send(ts.URL, []byte("some content"))

			a.CheckWantedErr(err, tc.wantErr)
		})
	}
}

func TestSendNoServer(t *testing.T) {
	t.Parallel()
	a := helper.Asserter{T: t}

	err := sender.Send("https://localhost:4299", []byte("some content"))

	a.CheckWantedErr(err, true)
}

type statusHandler int

func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(int(*h))
}
