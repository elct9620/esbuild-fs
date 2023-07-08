package esbuildfs

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

type PluginOptions struct {
	Outdir          string
	Prefix          string
	ServerSentEvent *ServerSentEventHandler
}

func Plugin(fsys *FS, options PluginOptions) (api.Plugin, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return api.Plugin{}, err
	}

	basePath := filepath.Join(cwd, options.Outdir)

	return api.Plugin{
		Name: "esbuild-fs",
		Setup: func(build api.PluginBuild) {
			build.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
				changes := make([]string, 0)
				for _, file := range result.OutputFiles {
					path, err := writeFile(fsys, basePath, options.Prefix, file)
					if err != nil {
						return api.OnEndResult{}, err
					}

					changes = append(changes, path)
				}

				if options.ServerSentEvent != nil {
					err := options.ServerSentEvent.NotifyChanged(changes)
					if err != nil {
						return api.OnEndResult{}, err
					}
				}

				return api.OnEndResult{}, nil
			})
		},
	}, nil
}

func writeFile(fsys *FS, basePath, prefix string, file api.OutputFile) (string, error) {
	path, err := getRelPath(basePath, prefix, file.Path)
	if err != nil {
		return "", err
	}

	return path, fsys.Write(path, bytes.NewReader(file.Contents))
}

func getRelPath(basePath, prefix, path string) (string, error) {
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return "", err
	}

	return filepath.Join(prefix, relPath), nil
}
