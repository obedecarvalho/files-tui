[default]
default:
    @just --list

# Compila o aplicativo em bin/files-tui.
build:
    go build -o bin/files-tui .

# Executa os testes.
test:
    go test ./...

# Executa a análise estática do Go.
vet:
    go vet ./...

# Formata os arquivos Go.
fmt:
    gofmt -w *.go

# Verifica a formatação sem alterar arquivos.
fmt-check:
    @files="$(gofmt -l *.go)"; if [ -n "$files" ]; then printf 'Arquivos não formatados:\n%s\n' "$files"; exit 1; fi

# Executa as verificações do projeto.
check: fmt-check vet test

# Executa o aplicativo, aceitando opções e um diretório.
run *args:
    go run . "$@"
set positional-arguments
