# Documentação do Fastfile

## Uso e opções

Para abrir outro diretório, passe seu caminho como argumento, por exemplo `./fastfile /home`.

Use `./fastfile --help` para exibir a versão e as opções disponíveis.

Sem argumento, abre o diretório atual. A navegação não altera arquivos nem executa arquivos selecionados; o atalho `e` abre um editor que pode alterá-los.

Por padrão, não é possível voltar acima do diretório inicial. Use `./fastfile --allow-parent /home` para permitir navegar até os diretórios pais. Essa limitação é de navegação, não uma barreira de segurança contra links simbólicos.

Pressione `Espaço` para sair e escrever o caminho absoluto do item em foco na saída padrão. Se não houver item selecionado, é escrito o caminho do diretório atual. Ao sair com `q` ou `Ctrl+C`, nenhum caminho é escrito.

Use `--reload` para atualizar automaticamente a listagem e, quando ativados, os estados Git a cada 5 segundos, por exemplo `./fastfile --reload /home`.

Os indicadores Git ficam desativados por padrão. Use `--git` para ativá-los ao iniciar ou pressione `Ctrl+g` para alterná-los durante a navegação.

Consulte os [atalhos no README](README.md#controles).

Os programas externos usam o terminal e, ao terminar, retornam à interface. Os atalhos `v` e `e` usam os comandos configurados em `fastfile.conf` ou, na ausência deles, `$PAGER` e `$EDITOR`, respectivamente. Se não estiverem definidos, uma mensagem é exibida. Visualização e edição exigem um arquivo comum. A listagem e os indicadores Git ativos são atualizados no retorno.

O atalho `d` funciona mesmo sem `--git` e abre com `git difftool` as alterações staged (índice em relação ao HEAD) e, em seguida, as não staged (arquivos em relação ao índice), também para diretórios. O Git abre a ferramenta para cada arquivo alterado; arquivos não rastreados não são comparados. Sem alterações rastreadas, nenhuma comparação é aberta. Fora de um repositório, é exibida uma mensagem. A comparação entre dois arquivos arbitrários não está disponível.

A prévia mostra metadados e até 256 KiB de texto UTF-8. Links simbólicos são marcados com `@`; diretórios, com `/`. Arquivos binários ou texto em outra codificação mostram apenas metadados.

Em repositórios Git, `[A]` indica arquivo novo no staging, `[S]` indica outras alterações staged, `[M]` indica alterações não staged e `[?]` indica arquivos não rastreados. Os marcadores podem ser combinados, como `[AM]`; diretórios agregam os estados dos itens contidos neles. As cores também indicam o estado, mas o item em foco mantém o destaque de seleção. Pressione `r` para atualizar os estados. Fora de um repositório ou sem Git instalado, a navegação funciona sem indicadores Git. Arquivos removidos não aparecem na listagem, mas suas pastas existentes recebem o indicador correspondente.

O marcador `[I]` identifica arquivos e diretórios ignorados pelo Git, diferenciando-os de arquivos novos (`[?]`) e modificados (`[M]`). Itens dentro de um diretório ignorado também recebem `[I]`; esse estado não é agregado às pastas ancestrais.

### Configuração de programas externos

Crie `$HOME/.config/fastfile.conf` para definir os comandos usados pelos atalhos `e`, `v` e `d`:

```ini
# Cada comando pode incluir argumentos
editor = vim
pager = bat --paging=always
git_diff = meld
```

Linhas vazias e linhas iniciadas por `#` são ignoradas. Os comandos de editor e visualizador aceitam argumentos, aspas e escapes, sem expansão ou execução pelo shell. Configurações vazias ou ausentes para `editor` e `pager` usam `$EDITOR` e `$PAGER`; para `git_diff`, uma mensagem é exibida se não estiver configurado. O arquivo é lido novamente ao acionar cada atalho.

Em `git_diff`, informe apenas o nome de uma ferramenta reconhecida ou configurada no Git, como `meld` ou `vimdiff` (consulte `git difftool --tool-help`). Para ferramentas personalizadas, configure `difftool.<nome>.cmd` no Git. Os comandos de editor e visualizador recebem o caminho do arquivo após `--`.

### Exemplos com variáveis de ambiente

Configure um visualizador para o atalho `v`, por exemplo:

```sh
export PAGER='less'
```

Ou, para usar o `bat`:

```sh
export PAGER='bat --paging=always'
```

Configure um editor para o atalho `e`, por exemplo:

```sh
export EDITOR='vim'
```

Depois, execute `./fastfile` no mesmo terminal.
