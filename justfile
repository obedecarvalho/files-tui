[default]
default:
    @just --list

# Build the application at bin/fastfile.
build:
    go build -o bin/fastfile .

# Run the tests.
test:
    go test ./...

# Run Go static analysis.
vet:
    go vet ./...

# Format the Go files.
fmt:
    gofmt -w *.go

# Check formatting without changing files.
fmt-check:
    @files="$(gofmt -l *.go)"; if [ -n "$files" ]; then printf 'Unformatted files:\n%s\n' "$files"; exit 1; fi

# Run project checks.
check: fmt-check vet test

# Run the application with optional flags and a directory.
run *args:
    go run . "$@"
set positional-arguments
