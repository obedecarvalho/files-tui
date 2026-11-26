# Files TUI

Navegador de arquivos para Linux em Go, com colunas Miller e prévia de arquivos.

## Instalação e execução

Os binários para Linux AMD64 e ARM64 são publicados nas releases do GitHub, com os nomes `files-tui-linux-amd64` e `files-tui-linux-arm64`. Baixe o arquivo correspondente à sua arquitetura, dê permissão de execução e execute-o, por exemplo:

```sh
chmod +x files-tui-linux-amd64
mv files-tui-linux-amd64 files-tui
./files-tui
```

Para compilar o projeto ou contribuir com o desenvolvimento, consulte [DEV.md](DEV.md).

Para abrir outro diretório, passe seu caminho como argumento, por exemplo `./files-tui /home`.

Sem argumento, abre o diretório atual. A interface não altera arquivos nem executa arquivos selecionados.

Por padrão, não é possível voltar acima do diretório inicial. Use `./files-tui --no-root /home` para permitir navegar até os diretórios pais. Essa limitação é de navegação, não uma barreira de segurança contra links simbólicos.

Ao sair, o caminho absoluto do item selecionado é escrito na saída padrão. Se não houver item selecionado, é escrito o caminho do diretório atual. Use `--quiet` para desativar essa saída.

Use `--reload` para atualizar automaticamente a listagem e, quando ativados, os estados Git a cada 5 segundos, por exemplo `./files-tui --reload /home`.

Os indicadores Git ficam desativados por padrão. Use `--git` para ativá-los ao iniciar ou pressione `Ctrl+g` para alterná-los durante a navegação.

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
