package esbuildfs

import (
	"net/http"

	"github.com/evanw/esbuild/pkg/api"
)

type PluginOptionFn func(options *PluginOptions)

func WithHandlerPrefix(prefix string) PluginOptionFn {
	return func(options *PluginOptions) {
		options.Prefix = prefix
	}
}

func Serve(buildOptions api.BuildOptions, options ...PluginOptionFn) (http.Handler, http.Handler, error) {
	fsys := New()
	sse := NewSSE()

	pluginOptions := PluginOptions{Outdir: buildOptions.Outdir, ServerSentEvent: sse, FileSystem: fsys}
	for _, fn := range options {
		fn(&pluginOptions)
	}

	plugin, err := Plugin(pluginOptions)
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
