ESBuild FS
===

This plugin creates a simple in-memory Key-Value filesystem to track compiled assets to provide `http.FileSystem` support.

## Usage

The `esbuildfs.Serve()` shortcut

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
