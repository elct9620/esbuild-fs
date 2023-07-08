package esbuildfs_test

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	esbuildfs "github.com/elct9620/esbuild-fs"
)

func Test_SSE(t *testing.T) {
	t.Parallel()

	fsys := esbuildfs.New()
	sse := esbuildfs.NewSSE()
	sse.Watch(fsys)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil).Clone(ctx)

	go sse.ServeHTTP(w, req)
	time.Sleep(1 * time.Millisecond)
	fsys.Write("app.js", bytes.NewBufferString(""))
	<-req.Context().Done()

	rawBody, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatal("unable to read SSE result")
	}

	body := string(rawBody)
	expected := `{"updated":["app.js"]}`
	if !strings.Contains(body, expected) {
		t.Fatal("Server Sent Event ", body, "not contains", expected)
	}
}
