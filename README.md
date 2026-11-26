# Files TUI

Navegador de arquivos para Linux em Go, com colunas Miller e prévia de arquivos.

## Compilar e executar

Requisitos: Linux, Go 1.26.4 ou posterior e acesso à internet para baixar as dependências na primeira compilação. Para usar as receitas abaixo, instale também o `just`. Na raiz do projeto, execute:

```sh
just build
./bin/files-tui
```

Para abrir outro diretório, passe seu caminho como argumento, por exemplo `./bin/files-tui /home`. Também é possível executar sem gerar um binário com `just run` ou `just run /home`.

Sem argumento, abre o diretório atual. A interface não altera arquivos nem executa arquivos selecionados.

Por padrão, não é possível voltar acima do diretório inicial. Use `./bin/files-tui --no-root /home` para permitir navegar até os diretórios pais. Essa limitação é de navegação, não uma barreira de segurança contra links simbólicos.

Ao sair, o caminho absoluto do item selecionado é escrito na saída padrão. Se não houver item selecionado, é escrito o caminho do diretório atual. Use `--quiet` para desativar essa saída.

Use `--reload` para atualizar automaticamente a listagem e, quando ativados, os estados Git a cada 5 segundos, por exemplo `./bin/files-tui --reload /home`.

Os indicadores Git ficam desativados por padrão. Use `--git` para ativá-los ao iniciar ou pressione `Ctrl+g` para alterná-los durante a navegação.

## Testes

Na raiz do projeto, execute todos os testes com:

```sh
just test
```

Para ver os testes individualmente, use `go test -v ./...`. Use `just vet` para a análise estática, `just fmt` para formatar o código e `just check` para verificar a formatação, executar a análise estática e rodar os testes. Execute `just --list` para listar as receitas disponíveis.

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
| Ctrl+g | Alternar indicadores Git |
| q ou Ctrl+C | Sair |

A prévia mostra metadados e até 256 KiB de texto UTF-8. Links simbólicos são marcados com `@`; diretórios, com `/`. Arquivos binários ou texto em outra codificação mostram apenas metadados.

Em repositórios Git, `[A]` indica arquivo novo no staging, `[S]` indica outras alterações staged, `[M]` indica alterações não staged e `[?]` indica arquivos não rastreados. Os marcadores podem ser combinados, como `[AM]`; diretórios agregam os estados dos itens contidos neles. As cores também indicam o estado, mas o item em foco mantém o destaque de seleção. Pressione `r` para atualizar os estados. Fora de um repositório ou sem Git instalado, a navegação funciona sem indicadores Git. Arquivos removidos não aparecem na listagem, mas suas pastas existentes recebem o indicador correspondente.

O marcador `[I]` identifica arquivos e diretórios ignorados pelo Git, diferenciando-os de arquivos novos (`[?]`) e modificados (`[M]`). Itens dentro de um diretório ignorado também recebem `[I]`; esse estado não é agregado às pastas ancestrais.
