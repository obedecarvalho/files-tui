# Files TUI

Navegador de arquivos para Linux em Go, com colunas Miller e prévia de arquivos.

## Instalação e execução

Baixe o arquivo mais recente:

```sh
curl -fL -o files-tui https://github.com/obedecarvalho/files-tui/releases/latest/download/files-tui-linux-amd64
```

```sh
chmod +x files-tui
./files-tui
```

Para compilar o projeto ou contribuir com o desenvolvimento, consulte [DEV.md](DEV.md).

Para abrir outro diretório, passe seu caminho como argumento, por exemplo `./files-tui /home`.

Use `./files-tui --help` para exibir a versão e as opções disponíveis.

Sem argumento, abre o diretório atual. A navegação não altera arquivos nem executa arquivos selecionados; o atalho `e` abre um editor que pode alterá-los.

Por padrão, não é possível voltar acima do diretório inicial. Use `./files-tui --no-root /home` para permitir navegar até os diretórios pais. Essa limitação é de navegação, não uma barreira de segurança contra links simbólicos.

Pressione `Espaço` para sair e escrever o caminho absoluto do item em foco na saída padrão. Se não houver item selecionado, é escrito o caminho do diretório atual. Ao sair com `q` ou `Ctrl+C`, nenhum caminho é escrito.

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
| v | Visualizar arquivo com $PAGER |
| e | Editar arquivo com $EDITOR |
| d | Mostrar diferenças Git do item selecionado |
| Ctrl+g | Alternar indicadores Git |
| ? | Abrir ou fechar a lista de atalhos; Esc também fecha |
| Espaço | Selecionar o item em foco e sair |
| q ou Ctrl+C | Sair |

Na lista de atalhos, use ↑/↓, k/j ou Page Up/Page Down para rolar.

Os programas externos usam o terminal e, ao terminar, retornam à interface. Os atalhos `v` e `e` exigem, respectivamente, que `$PAGER` e `$EDITOR` estejam definidos; caso contrário, uma mensagem é exibida. As variáveis aceitam argumentos e aspas, sem expansão de shell. Visualização e edição exigem um arquivo comum. A listagem e os indicadores Git ativos são atualizados no retorno.

O atalho `d` funciona mesmo sem `--git` e mostra separadamente alterações staged (índice em relação ao HEAD) e não staged (arquivos em relação ao índice), também para diretórios. Arquivos não rastreados são indicados, mas não aparecem no diff. Fora de um repositório, é exibida uma mensagem. A comparação entre dois arquivos arbitrários não está disponível.

A prévia mostra metadados e até 256 KiB de texto UTF-8. Links simbólicos são marcados com `@`; diretórios, com `/`. Arquivos binários ou texto em outra codificação mostram apenas metadados.

Em repositórios Git, `[A]` indica arquivo novo no staging, `[S]` indica outras alterações staged, `[M]` indica alterações não staged e `[?]` indica arquivos não rastreados. Os marcadores podem ser combinados, como `[AM]`; diretórios agregam os estados dos itens contidos neles. As cores também indicam o estado, mas o item em foco mantém o destaque de seleção. Pressione `r` para atualizar os estados. Fora de um repositório ou sem Git instalado, a navegação funciona sem indicadores Git. Arquivos removidos não aparecem na listagem, mas suas pastas existentes recebem o indicador correspondente.

O marcador `[I]` identifica arquivos e diretórios ignorados pelo Git, diferenciando-os de arquivos novos (`[?]`) e modificados (`[M]`). Itens dentro de um diretório ignorado também recebem `[I]`; esse estado não é agregado às pastas ancestrais.
