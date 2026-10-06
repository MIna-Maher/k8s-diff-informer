package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReadinessLifecycle(t *testing.T) {
	synced := false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := NewServer(8080, func() bool { return synced && ctx.Err() == nil })
	check := func(want int) {
		t.Helper()
		response := httptest.NewRecorder()
		server.readinessHandler(response, httptest.NewRequest("GET", "/ready", nil))
		if response.Code != want {
			t.Fatalf("got %d, want %d", response.Code, want)
		}
	}
	check(http.StatusServiceUnavailable)
	synced = true
	check(http.StatusOK)
	cancel()
	check(http.StatusServiceUnavailable)
	response := httptest.NewRecorder()
	server.healthHandler(response, httptest.NewRequest("GET", "/health", nil))
	if response.Code != http.StatusOK {
		t.Fatal("liveness must not depend on cache synchronization")
	}
}

func TestStartReportsBindFailure(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	server := NewServer(listener.Addr().(*net.TCPAddr).Port, nil)
	if err := server.Start(ctx); err == nil {
		t.Fatal("expected bind error")
	}
	if ctx.Err() != nil {
		t.Fatal("server waited for cancellation instead of returning bind error")
	}
}

func TestStartWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NewServer(0, nil).Start(ctx); err != nil {
		t.Fatal(err)
	}
}
