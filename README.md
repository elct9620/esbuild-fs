ESBuild FS
===
[![Test](https://github.com/elct9620/esbuild-fs/actions/workflows/test.yml/badge.svg)](https://github.com/elct9620/esbuild-fs/actions/workflows/test.yml)

This plugin creates a simple in-memory Key-Value filesystem to track compiled assets to provide `http.FileSystem` support.

> This is not designed for production, use it when development.

## Usage

Install with `go get`

```bash
go get -u github.com/elct9620/esbuild-fs
```

The `esbuildfs.Serve()` shortcut is created to make assets handler and SSE (Server Sent Event).

```go
assets, sse, err := esbuildfs.Server(
    api.BuildOptions{
        Outdir: "static/js", // required
        // ...
    },
    esbuildfs.WithHandlerPrefix("assets"),
)

// ...
mux.Handle("/assets/", assets)
mux.Handle("/esbuild", sse)
```

We can use `html/template` to add the Live Reload script.

```html
{{ if .LiveReload }}
<script>
    new EventSource('/esbuild').addEventListener('change', () => location.reload())
</script>
{{ end }}
```
