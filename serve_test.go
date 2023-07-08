package esbuildfs_test

import (
	"bufio"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	esbuildfs "github.com/elct9620/esbuild-fs"
	"github.com/evanw/esbuild/pkg/api"
)

func Test_Serve_Assets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name   string
		Prefix string
	}{
		{
			Name:   "server on root",
			Prefix: "",
		},
		{
			Name:   "with prefix",
			Prefix: "assets",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			options := givenAStdinBuild(t, []api.Plugin{}, "console.log(true)")
			assets, _ := givenAServeHandlers(t, options, esbuildfs.WithPrefix(tc.Prefix))
			serverURL := givenAServer(t, assets)
			ctx := givenAContextWithTimeout(t, 1*time.Second)

			for {
				select {
				case <-ctx.Done():
					t.Fatal("unable to find assets")
					return
				default:
					res := whenGetAssets(t, serverURL, filepath.Join(tc.Prefix, "stdin.js"))
					if thenMayFound(t, res) {
						return
					}
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
