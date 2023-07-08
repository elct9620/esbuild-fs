package esbuildfs

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

type PluginOptions struct {
	Outdir          string
	Prefix          string
	FileSystem      *FS
	ServerSentEvent *ServerSentEventHandler
}

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

func newPlugin(options PluginOptions) (*fsPlugin, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &fsPlugin{
		basePath: filepath.Join(cwd, options.Outdir),
		prefix:   options.Prefix,
		writer:   options.FileSystem,
		notifier: options.ServerSentEvent,
	}, nil
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

	return p.notifier.NotifyChanged(changes)
}

func (p *fsPlugin) Write(file api.OutputFile) (string, error) {
	path, err := p.RelPath(file.Path)
	if err != nil {
		return path, err
	}

	return path, p.writer.Write(path, bytes.NewBuffer(file.Contents))
}

func (p *fsPlugin) RelPath(path string) (string, error) {
	relPath, err := filepath.Rel(p.basePath, path)
	if err != nil {
		return path, err
	}

	return filepath.Join(p.prefix, relPath), nil
}

func Plugin(options PluginOptions) (api.Plugin, error) {
	plugin, err := newPlugin(options)
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
