# Development

## Requirements

Linux, Go 1.26.4 or later, and internet access to download dependencies on the first build. Install `just` to use the recipes below.

## Build and run

From the project root, run:

```sh
just build
./bin/fastfile
```

You can also run without building a binary using `just run` or `just run /home`.
Local builds show the version `dev` in `--help`; release versions match the tag.

## Tests and checks

Run all tests with `just test` or see individual tests with `go test -v ./...`. Use `just vet` for static analysis, `just fmt` to format the code, and `just check` to check formatting, run static analysis and run tests. Run `just --list` to list the available recipes.

## Releases

When a new tag is pushed, GitHub Actions checks formatting, runs static analysis and tests, builds Linux AMD64 and ARM64 binaries, and creates a release with them.
