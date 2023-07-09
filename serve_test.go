package esbuildfs_test

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	esbuildfs "github.com/elct9620/esbuild-fs"
	"github.com/evanw/esbuild/pkg/api"
)

func Test_Serve_Assets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name         string
		MountTo      string
		Prefix       string
		Wrapper      func(http.Handler) http.Handler
		ExpectedPath string
	}{
		{
			Name:    "server on root",
			MountTo: "/",
			Prefix:  "",
			Wrapper: func(handler http.Handler) http.Handler {
				return handler
			},
			ExpectedPath: "stdin.js",
		},
		{
			Name:    "with prefix",
			MountTo: "/",
			Prefix:  "assets",
			Wrapper: func(handler http.Handler) http.Handler {
				return handler
			},
			ExpectedPath: "assets/stdin.js",
		},
		{
			Name:    "with prefix use http.StripPrefix",
			MountTo: "/assets/",
			Prefix:  "",
			Wrapper: func(handler http.Handler) http.Handler {
				return http.StripPrefix("/assets", handler)
			},
			ExpectedPath: "assets/stdin.js",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			options := givenAStdinBuild(t, []api.Plugin{}, "console.log(true)")
			assets, _ := givenAServeHandlers(t, options, esbuildfs.WithPrefix(tc.Prefix))
			serverURL := givenAServerMountTo(t, tc.MountTo, tc.Wrapper(assets))
			ctx := givenAContextWithTimeout(t, 5*time.Second)

			for {
				select {
				case <-ctx.Done():
					t.Fatal("unable to find assets", tc.ExpectedPath)
					return
				default:
					res := whenGetAssets(t, serverURL, tc.ExpectedPath)
					if thenMayFound(t, res) {
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		})
	}
}

func Test_Serve_SSE(t *testing.T) {
	options := givenAStdinBuild(t, []api.Plugin{}, "console.log(true)")
	_, sse := givenAServeHandlers(t, options)
	serverURL := givenAServer(t, sse)
	ctx := givenAContextWithTimeout(t, 1*time.Second)
	res := whenConnectSSE(t, ctx, serverURL)
	events := whenReceiveEvents(t, bufio.NewReader(res.Body))
	thenCanSeeEvent(t, ctx, events, "retry")
}

func givenAServerMountTo(t *testing.T, path string, handler http.Handler) string {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(func() {
		server.Close()
	})

	return server.URL
}

func givenAServeHandlers(t *testing.T, buildOptions api.BuildOptions, options ...esbuildfs.PluginOptionFn) (http.Handler, http.Handler) {
	t.Helper()

	assets, sse, err := esbuildfs.Serve(buildOptions, options...)
	if err != nil {
		t.Fatal("unable to serve esbuild", err)
	}

	return assets, sse
}

func whenGetAssets(t *testing.T, serverURL, name string) *http.Response {
	t.Helper()

	res, err := http.Get(fmt.Sprintf("%s/%s", serverURL, name))
	if err != nil {
		t.Fatal("unable to get assets", err)
	}

	return res
}

func thenMayFound(t *testing.T, res *http.Response) bool {
	t.Helper()

	return res.StatusCode == http.StatusOK
}
