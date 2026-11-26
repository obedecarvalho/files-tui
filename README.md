# Files TUI

Navegador de arquivos para Linux em Go, com colunas Miller e prévia de arquivos.

## Compilar e executar

Requisitos: Linux, Go 1.26.4 ou posterior e acesso à internet para baixar as dependências na primeira compilação. Na raiz do projeto, execute:

```sh
go build -o files-tui .
./files-tui
```

Para abrir outro diretório, passe seu caminho como argumento, por exemplo `./files-tui /home`. Também é possível executar sem gerar um binário com `go run .` ou `go run . /home`.

Sem argumento, abre o diretório atual. A interface não altera arquivos nem executa arquivos selecionados.

Por padrão, não é possível voltar acima do diretório inicial. Use `./files-tui --no-root /home` para permitir navegar até os diretórios pais. Essa limitação é de navegação, não uma barreira de segurança contra links simbólicos.

## Testes

Na raiz do projeto, execute todos os testes com:

```sh
go test ./...
```

Para ver os testes individualmente, use `go test -v ./...`. Para executar também a análise estática do Go, use `go vet ./...`.

## Controles

| Tecla | Ação |
| --- | --- |
| ↑/↓ ou k/j | Mover seleção |
| →, l ou Enter | Entrar no diretório |
| ←, h ou Backspace | Voltar ao diretório pai |
| g / G | Primeiro / último item |
| [ / ] | Deslocar colunas manualmente |
| Page Up / Page Down | Rolar prévia em até três linhas |
| . | Alternar arquivos ocultos |
| / | Filtrar nomes; Enter confirma e Esc limpa |
| s | Alternar ordenação por nome, tamanho e data |
| r | Atualizar listagem |
| q ou Ctrl+C | Sair |

A prévia mostra metadados e até 256 KiB de texto UTF-8. Links simbólicos são marcados com `@`; diretórios, com `/`. Arquivos binários ou texto em outra codificação mostram apenas metadados.
