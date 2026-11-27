# Fastfile

Navegador de arquivos para Linux em Go, com colunas Miller e prévia de arquivos.

## Instalação e execução

Para baixar a release mais recente para Linux AMD64:

```sh
VERSION=$(curl -fsSL https://api.github.com/repos/obedecarvalho/fastfile/releases/latest | grep '"tag_name":' | cut -d '"' -f 4)
curl -fL -o fastfile "https://github.com/obedecarvalho/fastfile/releases/download/${VERSION}/fastfile-linux-${VERSION}-amd64"
```

Para Linux ARM64, substitua `amd64` por `arm64` no comando de download.

```sh
chmod +x fastfile
./fastfile
```

Para abrir outro diretório, execute `./fastfile /home`. Use `./fastfile --help` para ver as opções disponíveis e `?` durante a navegação para consultar os atalhos.

Para executar `fastfile` de qualquer diretório, instale-o em `$HOME/.local/bin`:

```sh
mkdir -p "$HOME/.local/bin"
mv fastfile "$HOME/.local/bin/fastfile"
```

Se `$HOME/.local/bin` ainda não estiver no `PATH`, adicione a linha abaixo ao arquivo de configuração do seu shell (por exemplo, `~/.bashrc` ou `~/.zshrc`) e abra um novo terminal:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Então execute `fastfile` ou `fastfile /home`.

## Controles

| Tecla | Ação |
| --- | --- |
| ↑/↓ ou k/j | Mover seleção |
| →, l ou Enter | Entrar no diretório |
| ←, h ou Backspace | Voltar ao diretório pai |
| g / G | Primeiro / último item |
| [ / ] | Deslocar colunas manualmente |
| Page Up / Page Down | Rolar prévia em até três linhas |
| p | Mostrar ou ocultar a prévia |
| P / Esc | Ampliar ou restaurar a prévia |
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

Configure editor, visualizador e diff Git em `$HOME/.config/fastfile.conf`; veja os [detalhes de configuração](DOCS.md#configuração-de-programas-externos).

## Documentação

- [Uso, controles e funcionalidades](DOCS.md)
- [Compilação e desenvolvimento](DEV.md)
