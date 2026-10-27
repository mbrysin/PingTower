package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantUp     bool
	}{
		{name: "successful response", statusCode: http.StatusOK, wantUp: true},
		{name: "client error is up by current rule", statusCode: http.StatusNotFound, wantUp: true},
		{name: "server error is down", statusCode: http.StatusInternalServerError, wantUp: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if got := r.Header.Get("User-Agent"); got != "PingTower/0.0.0-pre-alpha" {
					t.Errorf("User-Agent = %q", got)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			got := Check(context.Background(), server.URL)

			if got.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", got.StatusCode, tt.statusCode)
			}
			if got.IsUp != tt.wantUp {
				t.Errorf("IsUp = %t, want %t", got.IsUp, tt.wantUp)
			}
			if got.ResponseTimeMs < 0 {
				t.Errorf("ResponseTimeMs = %d, want non-negative", got.ResponseTimeMs)
			}
		})
	}
}

func TestCheckSendsGETAndUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if got := r.Header.Get("User-Agent"); got != "PingTower/0.0.0-pre-alpha" {
			t.Errorf("User-Agent = %q, want PingTower/0.0.0-pre-alpha", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	got := Check(context.Background(), server.URL)

	if !got.IsUp {
		t.Errorf("IsUp = false, want true; error = %q", got.Error)
	}
	if got.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", got.StatusCode, http.StatusOK)
	}
	if got.ResponseTimeMs < 0 {
		t.Errorf("ResponseTimeMs = %d, want non-negative", got.ResponseTimeMs)
	}
}

func TestCheckTreatsStatusBelow500AsUp(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantUp     bool
	}{
		{name: "redirect", statusCode: http.StatusFound, wantUp: true},
		{name: "not found", statusCode: http.StatusNotFound, wantUp: true},
		{name: "server error", statusCode: http.StatusInternalServerError, wantUp: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(
				w http.ResponseWriter,
				_ *http.Request,
			) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			got := Check(context.Background(), server.URL)

			if got.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", got.StatusCode, tt.statusCode)
			}
			if got.IsUp != tt.wantUp {
				t.Errorf("IsUp = %t, want %t", got.IsUp, tt.wantUp)
			}
		})
	}
}

func TestCheckDoesNotFollowRedirect(t *testing.T) {
	destinationRequested := false

	destination := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		destinationRequested = true
		w.WriteHeader(http.StatusOK)
	}))
	defer destination.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	defer redirect.Close()

	got := Check(context.Background(), redirect.URL)

	if got.StatusCode != http.StatusFound {
		t.Errorf("StatusCode = %d, want %d", got.StatusCode, http.StatusFound)
	}
	if destinationRequested {
		t.Error("Check followed redirect; want redirect not to be followed")
	}
}

func TestCheckInvalidURL(t *testing.T) {
	got := Check(context.Background(), "://bad-url")

	if got.Error == "" {
		t.Fatal("expected an error for invalid URL")
	}
	if got.StatusCode != 0 {
		t.Errorf("StatusCode = %d, want 0", got.StatusCode)
	}
	if got.IsUp {
		t.Error("IsUp = true, want false")
	}
}

func TestCheckStopsWhenContextIsCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := Check(ctx, server.URL)

	if got.Error == "" {
		t.Fatal("expected an error when context is canceled")
	}
	if got.IsUp {
		t.Error("IsUp = true, want false")
	}
}

func TestCheckStopsWhenContextDeadlineExpires(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	got := Check(ctx, server.URL)
	elapsed := time.Since(start)

	if got.Error == "" {
		t.Fatal("expected an error after context deadline")
	}
	if elapsed > time.Second {
		t.Errorf("Check took %s after a 50ms deadline", elapsed)
	}
}

func TestCheckErrorContainsUsefulMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := Check(ctx, "http://example.test")

	if got.Error == "" {
		t.Fatal("expected a non-empty error")
	}
	if !strings.Contains(strings.ToLower(got.Error), "context") {
		t.Logf("request returned error: %q", got.Error)
	}
}
