# Fastfile

A Linux file browser written in Go, with Miller columns and file previews.

## Installation and usage

To download the latest release for Linux AMD64:

```sh
VERSION=$(curl -fsSL https://api.github.com/repos/obedecarvalho/fastfile/releases/latest | grep '"tag_name":' | cut -d '"' -f 4)
curl -fL -o fastfile "https://github.com/obedecarvalho/fastfile/releases/download/${VERSION}/fastfile-linux-${VERSION}-amd64"
```

For Linux ARM64, replace `amd64` with `arm64` in the download command.

```sh
chmod +x fastfile
./fastfile
```

To open another directory, run `./fastfile /home`. Use `./fastfile --help` to see the available options and `?` while browsing to see the shortcuts.

To run `fastfile` from any directory, install it in `$HOME/.local/bin`:

```sh
mkdir -p "$HOME/.local/bin"
mv fastfile "$HOME/.local/bin/fastfile"
```

If `$HOME/.local/bin` is not already in your `PATH`, add the following line to your shell configuration file (for example, `~/.bashrc` or `~/.zshrc`) and open a new terminal:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Then run `fastfile` or `fastfile /home`.

## Controls

| Key | Action |
| --- | --- |
| ↑/↓ or k/j | Move selection |
| →, l or Enter | Enter directory |
| ←, h or Backspace | Go to parent directory |
| g / G | First / last item |
| [ / ] | Shift columns manually |
| Page Up / Page Down | Scroll preview by up to three lines |
| p | Show or hide preview |
| P / Esc | Expand or restore preview |
| . | Toggle hidden files |
| / | Filter names; Enter confirms and Esc clears |
| s | Change sort order between name, size and date |
| r | Refresh listing |
| v | View file with $PAGER |
| e | Edit file with $EDITOR |
| d | Show Git differences for the selected item |
| Ctrl+g | Toggle Git indicators |
| ? | Open or close shortcuts; Esc also closes them |
| Space | Select the focused item and exit |
| q or Ctrl+C | Exit |

In the shortcuts list, use ↑/↓, k/j or Page Up/Page Down to scroll.

Configure the editor, pager and Git diff tool in `$HOME/.config/fastfile.conf`; see [configuration details](DOCS.md#external-program-configuration).

## Documentation

- [Usage, controls and features](DOCS.md)
- [Building and development](DEV.md)
