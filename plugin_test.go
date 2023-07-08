package esbuildfs_test

import (
	"io/fs"
	"testing"

	esbuildfs "github.com/elct9620/esbuild-fs"
	"github.com/evanw/esbuild/pkg/api"
)

func Test_Plugin(t *testing.T) {
	fsys := esbuildfs.New()
	plugin := givenAnNewPlugin(t, fsys)
	options := givenAStdinBuild(t, []api.Plugin{plugin}, "console.log(true)")
	whenBuild(t, options)
	thenCanOpenFile(t, fsys, "stdin.js")
}

func givenAnNewPlugin(t *testing.T, writer esbuildfs.Writer, options ...esbuildfs.PluginOptionFn) api.Plugin {
	t.Helper()

	plugin, err := esbuildfs.Plugin("dist", writer, options...)
	if err != nil {
		t.Fatal("unable to initialize plugin", err)
	}

	return plugin
}

func givenAStdinBuild(t *testing.T, plugins []api.Plugin, content string) api.BuildOptions {
	t.Helper()

	return api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents: content,
		},
		Outdir:   "dist",
		Plugins:  plugins,
		LogLevel: api.LogLevelDebug,
	}
}

func whenBuild(t *testing.T, options api.BuildOptions) {
	t.Helper()

	res := api.Build(options)
	if len(res.Errors) > 0 {
		t.Fatal("unable to build assets", res.Errors)
	}
}

func thenCanOpenFile(t *testing.T, fsys fs.FS, name string) {
	t.Helper()

	_, err := fsys.Open(name)
	if err != nil {
		t.Fatal("unable to open file", err)
	}
}
