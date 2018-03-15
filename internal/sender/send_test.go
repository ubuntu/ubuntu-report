package sender_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ubuntu/ubuntu-report/internal/sender"
)

func TestGetURL(t *testing.T) {
	t.Parallel()

	got, err := sender.GetURL("https://myurl.com", "distroname", "versionnumber")
	want := "https://myurl.com/distroname/desktop/versionnumber"

	if err != nil {
		t.Fatal("got a parsing error:", err)
	}
	if got != want {
		t.Errorf("got %s; want %s", got, want)
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

			status := statusHandler(tc.status)
			ts := httptest.NewServer(&status)
			defer ts.Close()

			err := sender.Send(ts.URL, []byte("some content"))
			if err != nil && !tc.wantErr {
				t.Fatal("got an unexpected err:", err)
			}
			if err == nil && tc.wantErr {
				t.Error("expected an error and got none")
			}
		})
	}
}

type statusHandler int

func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(int(*h))
}
