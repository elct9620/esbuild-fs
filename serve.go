package esbuildfs

import (
	"net/http"

	"github.com/evanw/esbuild/pkg/api"
)

func Serve(buildOptions api.BuildOptions, options ...PluginOptionFn) (http.Handler, http.Handler, error) {
	fsys := New()
	sse := NewSSE()
	options = append(options, WithNotifier(sse))
	plugin, err := Plugin(buildOptions.Outdir, fsys, options...)
	if err != nil {
		return nil, nil, err
	}

	buildOptions.Plugins = append(buildOptions.Plugins, plugin)
	ctx, ctxErr := api.Context(buildOptions)
	if ctxErr != nil {
		return nil, nil, ctxErr
	}

	err = ctx.Watch(api.WatchOptions{})
	if err != nil {
		return nil, nil, err
	}

	return http.FileServer(http.FS(fsys)), sse, nil
}
