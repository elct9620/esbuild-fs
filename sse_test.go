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
	server := httptest.NewServer(sse)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	res := connectSSE(t, ctx, server.URL)
	events := readEvents(bufio.NewReader(res.Body))

	err := sse.NotifyChanged([]string{"app.js"})
	if err != nil {
		t.Fatal("unable to notify changes", err)
	}

	expected := `{"updated":["app.js"]}`
	for {
		select {
		case <-ctx.Done():
			t.Fatal("unable to find expected event", expected)
		case event := <-events:
			if strings.Contains(event, expected) {
				return
			}
		}
	}
}

func connectSSE(t *testing.T, ctx context.Context, url string) *http.Response {
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

func readEvents(reader *bufio.Reader) chan string {
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
