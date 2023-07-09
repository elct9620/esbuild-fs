package esbuildfs

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

type PluginOptionFn = func(*fsPlugin)

type Writer interface {
	Write(name string, content io.Reader) error
}

type Notifier interface {
	NotifyChanged(updated []string) error
}

type fsPlugin struct {
	basePath string
	prefix   string
	writer   Writer
	notifier Notifier
}

func newPlugin(outdir string, writer Writer, options ...PluginOptionFn) (*fsPlugin, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	plugin := &fsPlugin{
		basePath: formatPath(path.Join(cwd, outdir)),
		writer:   writer,
	}

	for _, fn := range options {
		fn(plugin)
	}

	return plugin, nil
}

func (p *fsPlugin) Update(files []api.OutputFile) error {
	changes := make([]string, 0)

	for _, file := range files {
		path, err := p.Write(file)
		if err != nil {
			return err
		}

		changes = append(changes, path)
	}

	if p.notifier == nil {
		return nil
	}

	return p.notifier.NotifyChanged(changes)
}

func (p *fsPlugin) Write(file api.OutputFile) (string, error) {
	path, err := p.RelPath(file.Path)
	if err != nil {
		return path, err
	}

	return path, p.writer.Write(path, bytes.NewBuffer(file.Contents))
}

func (p *fsPlugin) RelPath(name string) (string, error) {
	relPath, err := filepath.Rel(p.basePath, formatPath(name))
	if err != nil {
		return name, err
	}

	return path.Join(p.prefix, relPath), nil
}

func WithPrefix(prefix string) PluginOptionFn {
	return func(plugin *fsPlugin) {
		plugin.prefix = prefix
	}
}

func WithNotifier(notifier Notifier) PluginOptionFn {
	return func(plugin *fsPlugin) {
		plugin.notifier = notifier
	}
}

func Plugin(outdir string, writer Writer, options ...PluginOptionFn) (api.Plugin, error) {
	plugin, err := newPlugin(outdir, writer, options...)
	if err != nil {
		return api.Plugin{}, err
	}

	return api.Plugin{
		Name: "esbuild-fs",
		Setup: func(build api.PluginBuild) {
			build.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
				err := plugin.Update(result.OutputFiles)
				if err != nil {
					return api.OnEndResult{}, err
				}

				return api.OnEndResult{}, nil
			})
		},
	}, nil
}

func formatPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
