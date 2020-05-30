# BM25 implementation

## Build and run

```bash
$ go run cmd/app/main.go
```

## Config

Have a look at `.config/search_config.json`

## App behaviour 

The app writes the the built index to disk.

The app writes the current state to disk.

The paths can be configured in `.config/search_config.json`.

Rebuilding the index might or might not be faster than loading it from disk.

If there is a new file or an existing file is modified, the entire index is rebuild.