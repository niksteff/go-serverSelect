package serverselect_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	serverselect "github.com/niksteff/go-serverSelect"
)

func TestServerSelect(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/fast" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if path == "/slow" {
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusOK)
			return
		}
	}))
	defer testSrv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	winner := serverselect.Racer(ctx, fmt.Sprintf("%s/fast", testSrv.URL), fmt.Sprintf("%s/slow", testSrv.URL))
	t.Logf("Winner: %#v\n", winner)
	if strings.Compare(winner, fmt.Sprintf("%s/fast", testSrv.URL)) != 0 {
		t.Errorf("expected winner to be %q but got %q", fmt.Sprintf("%s/fast", testSrv.URL), winner)
		t.FailNow()
		return
	}
}

func TestServerSelectVariadic(t *testing.T) {
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/fast" {
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			return
		}

		if path == "/mid" {
			time.Sleep(150 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			return
		}

		if path == "/slow" {
			time.Sleep(500 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			return
		}
	}))
	defer testSrv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testTable := []struct {
		name           string
		urls           []string
		expectedWinner string
	}{
		{
			name: "fast and slow",
			urls: []string{
				fmt.Sprintf("%s/slow", testSrv.URL),
				fmt.Sprintf("%s/fast", testSrv.URL),
			},
			expectedWinner: fmt.Sprintf("%s/fast", testSrv.URL),
		},
		{
			name: "slow and slow",
			urls: []string{
				fmt.Sprintf("%s/slow", testSrv.URL),
				fmt.Sprintf("%s/slow", testSrv.URL),
			},
			expectedWinner: fmt.Sprintf("%s/slow", testSrv.URL),
		},
		{
			name: "fast and fast",
			urls: []string{
				fmt.Sprintf("%s/fast", testSrv.URL),
				fmt.Sprintf("%s/fast", testSrv.URL),
			},
			expectedWinner: fmt.Sprintf("%s/fast", testSrv.URL),
		},
		{
			name: "slow and mid and fast",
			urls: []string{
				fmt.Sprintf("%s/slow", testSrv.URL),
				fmt.Sprintf("%s/mid", testSrv.URL),
				fmt.Sprintf("%s/fast", testSrv.URL),
			},
			expectedWinner: fmt.Sprintf("%s/fast", testSrv.URL),
		},
		{
			name: "slow and slow and mid and fast",
			urls: []string{
				fmt.Sprintf("%s/slow", testSrv.URL),
				fmt.Sprintf("%s/slow", testSrv.URL),
				fmt.Sprintf("%s/mid", testSrv.URL),
				fmt.Sprintf("%s/fast", testSrv.URL),
			},
			expectedWinner: fmt.Sprintf("%s/fast", testSrv.URL),
		},
	}

	t.Parallel()
	for _, d := range testTable {
		t.Run(d.name, func(t *testing.T) {
			winner := serverselect.RacerVar(ctx, d.urls...)
			if strings.Compare(winner, d.expectedWinner) != 0 {
				t.Errorf("expected winner to be %q but got %q", d.expectedWinner, winner)
				return
			}

			t.Logf("pass: %q", d.name)
			return
		})
	}
}
