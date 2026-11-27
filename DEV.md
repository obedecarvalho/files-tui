# Desenvolvimento

## Requisitos

Linux, Go 1.26.4 ou posterior e acesso à internet para baixar as dependências na primeira compilação. Para usar as receitas abaixo, instale também o `just`.

## Compilar e executar

Na raiz do projeto, execute:

```sh
just build
./bin/fastfile
```

Também é possível executar sem gerar um binário com `just run` ou `just run /home`.
Compilações locais exibem a versão `dev` em `--help`; nas releases, a versão corresponde à tag.

## Testes e verificações

Execute todos os testes com `just test` ou veja os testes individualmente com `go test -v ./...`. Use `just vet` para a análise estática, `just fmt` para formatar o código e `just check` para verificar a formatação, executar a análise estática e rodar os testes. Execute `just --list` para listar as receitas disponíveis.

## Releases

Ao enviar uma nova tag, o GitHub Actions verifica a formatação, executa a análise estática e os testes, compila os binários para Linux AMD64 e ARM64 e cria uma release com eles.
