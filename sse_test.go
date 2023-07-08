package esbuildfs_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	esbuildfs "github.com/elct9620/esbuild-fs"
)

func Test_SSE(t *testing.T) {
	t.Parallel()

	sse := esbuildfs.NewSSE()
	serverURL := givenAServer(t, sse)
	ctx := givenAContextWithTimeout(t, 1*time.Second)

	res := whenConnectSSE(t, ctx, serverURL)
	events := whenReceiveEvents(t, bufio.NewReader(res.Body))
	whenNotifyChanges(t, sse, []string{"app.js"})
	thenCanSeeEvent(t, ctx, events, `{"updated":["app.js"]}`)
}

func givenAServer(t *testing.T, handler http.Handler) string {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})

	return server.URL
}

func givenAContextWithTimeout(t *testing.T, duration time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	t.Cleanup(func() {
		cancel()
	})

	return ctx
}

func whenConnectSSE(t *testing.T, ctx context.Context, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Fatal("unable to create http request", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("unable to connect SSE", err)
	}

	return res
}

func whenReceiveEvents(t *testing.T, reader *bufio.Reader) chan string {
	t.Helper()

	events := make(chan string)
	go func() {
		for {
			next, err := reader.ReadString('\n')
			if err != nil {
				close(events)
				break
			}

			events <- next
		}
	}()

	return events
}

func whenNotifyChanges(t *testing.T, notifier esbuildfs.Notifier, changes []string) {
	t.Helper()

	err := notifier.NotifyChanged(changes)
	if err != nil {
		t.Fatal("unable to notify changes", err)
	}
}

func thenCanSeeEvent(t *testing.T, ctx context.Context, events chan string, payload string) {
	t.Helper()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("unable to find expected event", payload)
		case event := <-events:
			if strings.Contains(event, payload) {
				return
			}
		}
	}
}
