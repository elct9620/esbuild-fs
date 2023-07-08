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

	whenBuildStdin(t, plugin, "console.log(true)")
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

func whenBuildStdin(t *testing.T, plugin api.Plugin, content string) {
	t.Helper()

	res := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents: content,
		},
		Outdir:  "dist",
		Plugins: []api.Plugin{plugin},
	})

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
