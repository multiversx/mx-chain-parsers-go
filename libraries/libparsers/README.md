# libparsers

Parsers of blockchain / indexer entities, compiled as a library (shared object).

## Build

On Linux:

```
go build -buildmode=c-shared -o libparsers.so .
```

On MacOS:

```
go build -buildmode=c-shared -o libparsers.dylib .
```
