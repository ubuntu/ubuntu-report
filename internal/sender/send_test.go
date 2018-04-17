package sender_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestSendInfiniteRequestServer(t *testing.T) {
	helper.SkipIfShort(t)
	t.Parallel()
	a := helper.Asserter{T: t}

	// not exactly an inifite request, but a long one
	closehandler := make(chan struct{})
	handlerclosed := make(chan struct{})
	timeout := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerclosed)
		select {
		case <-time.After(20 * time.Second):
			timeout = true
		case <-r.Context().Done():
		case <-closehandler:
		}
	}))
	defer ts.Close()

	err := sender.Send(ts.URL, []byte("some content"))

	// ensure we get the handler close to setup cancelled flag if timeout not reached
	close(closehandler)
	<-handlerclosed

	a.CheckWantedErr(err, true)
	if timeout {
		t.Errorf("Expected to let client cancelling server side answer and it didn't.")
	}
}

type statusHandler int

func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(int(*h))
}
