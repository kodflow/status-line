# Go >= 1.25.0
> Release Notes: https://go.dev/doc/devel/release

## Structure

```
/src
├── cmd/           # Entry points
│   └── app/
│       └── main.go
├── internal/      # Private packages
├── pkg/           # Public packages
├── go.mod
└── go.sum
```

**Tests:** Alongside code (`*_test.go`), NOT in `/tests`

## Style

- `gofmt` + `goimports` (mandatory)
- `golangci-lint` for linting
- Effective Go conventions
- No `init()` unless absolutely necessary
- Context as first parameter
- Errors as last return value

## Naming

- Packages: lowercase, single word
- Interfaces: `-er` suffix (Reader, Writer)
- Exported: PascalCase
- Unexported: camelCase

## Testing

- Table-driven tests
- `testify` for assertions
- `go test -race` for race detection
- `go test -cover` minimum 80%

## Modules

- Go modules (go.mod) mandatory
- Semantic versioning
- `go mod tidy` before commit

## Forbidden

- `panic` for error handling
- Global mutable state
- `interface{}` without type assertion
- Naked returns in long functions
