[default]
default:
    @just --list

build:
    go build -o files-tui .

test:
    go test ./...