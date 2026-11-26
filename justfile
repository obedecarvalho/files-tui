[default]
default:
    @just --list

build:
    go build -o bin/files-tui .

test:
    go test ./...

run:
    go run .
