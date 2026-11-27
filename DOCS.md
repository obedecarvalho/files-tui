# Fastfile documentation

## Usage and options

To open another directory, pass its path as an argument, for example `./fastfile /home`.

Use `./fastfile --help` to show the version and available options.

With no argument, Fastfile opens the current directory. Browsing does not change files or execute selected files; the `e` shortcut opens an editor that can change them.

By default, you cannot navigate above the starting directory. Use `./fastfile --allow-parent /home` to allow navigation to parent directories. This is a navigation restriction, not a security barrier against symbolic links.

Press `Space` to exit and write the absolute path of the focused item to standard output. If no item is selected, the current directory path is written instead. Exiting with `q` or `Ctrl+C` writes no path.

Use `--reload` to automatically refresh the listing and, when enabled, Git statuses every 5 seconds, for example `./fastfile --reload /home`.

Git indicators are disabled by default. Use `--git` to enable them at startup or press `Ctrl+g` to toggle them while browsing.

See the [shortcuts in the README](README.md#controls).

External programs use the terminal and return to the interface when they finish. The `v` and `e` shortcuts use commands configured in `fastfile.conf` or, if absent, `$PAGER` and `$EDITOR`, respectively. If they are not set, a message is displayed. Viewing and editing require a regular file. The listing and any active Git indicators are refreshed on return.

The `d` shortcut works even without `--git` and uses `git difftool` to open staged changes (index against HEAD) followed by unstaged changes (files against the index), including for directories. Git opens the tool for each changed file; untracked files are not compared. If there are no tracked changes, no comparison is opened. Outside a repository, a message is displayed. Comparing two arbitrary files is not available.

The preview is enabled by default and can be shown or hidden with `p` when there is enough screen space. Use `P` to expand the preview and press `P` again or `Esc` to return to the normal layout. It shows type, human-readable size, modification time, permissions and up to 256 KiB of UTF-8 text. Symbolic links are marked with `@` and show their target; directories are marked with `/`. For binary files or text in another encoding, it also shows the detected content type when possible.

In Git repositories, `[A]` indicates a newly staged file, `[S]` indicates other staged changes, `[M]` indicates unstaged changes and `[?]` indicates untracked files. Markers can be combined, such as `[AM]`; directories aggregate the statuses of items they contain. Colors also indicate status, but the focused item retains its selection highlight. Press `r` to refresh statuses. Outside a repository or without Git installed, browsing works without Git indicators. Deleted files do not appear in the listing, but their existing directories receive the corresponding indicator.

The `[I]` marker identifies files and directories ignored by Git, distinguishing them from new (`[?]`) and modified (`[M]`) files. Items inside an ignored directory also receive `[I]`; this status is not aggregated to ancestor directories.

### External program configuration

Create `$HOME/.config/fastfile.conf` to define the commands used by the `e`, `v` and `d` shortcuts:

```ini
# Each command may include arguments
editor = vim
pager = bat --paging=always
git_diff = meld
```

Empty lines and lines beginning with `#` are ignored. Editor and pager commands accept arguments, quotes and escapes, without shell expansion or execution. Empty or missing `editor` and `pager` settings fall back to `$EDITOR` and `$PAGER`; for `git_diff`, a message is displayed if it is not configured. The file is read again each time a shortcut is invoked.

For `git_diff`, specify only the name of a tool recognized or configured by Git, such as `meld` or `vimdiff` (see `git difftool --tool-help`). For custom tools, configure `difftool.<name>.cmd` in Git. Editor and pager commands receive the file path after `--`.

### Environment variable examples

Configure a pager for the `v` shortcut, for example:

```sh
export PAGER='less'
```

Or, to use `bat`:

```sh
export PAGER='bat --paging=always'
```

Configure an editor for the `e` shortcut, for example:

```sh
export EDITOR='vim'
```

Then run `./fastfile` in the same terminal.
