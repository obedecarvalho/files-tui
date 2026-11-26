package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

type externalFinished struct{ err error }

type externalTool struct{ action, path, dir string }

func externalCommand(action, path, dir string) tea.Cmd {
	return tea.Exec(&externalTool{action: action, path: path, dir: dir}, func(err error) tea.Msg {
		return externalFinished{err: err}
	})
}

// External programs use the controlling terminal, never the selection's stdout.
func (*externalTool) SetStdin(io.Reader)  {}
func (*externalTool) SetStdout(io.Writer) {}
func (*externalTool) SetStderr(io.Writer) {}

func (t *externalTool) Run() error {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("terminal indisponível: %w", err)
	}
	defer tty.Close()
	if t.action == "d" {
		return t.diff(tty)
	}
	variable := "PAGER"
	if t.action == "e" {
		variable = "EDITOR"
	}
	value := os.Getenv(variable)
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s não está definido", variable)
	}
	args, err := commandWords(value)
	if err != nil || len(args) == 0 || args[0] == "" {
		return fmt.Errorf("%s inválido", variable)
	}
	args = append(args, "--", t.path)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", args[0], err)
	}
	if t.action == "v" && filepath.Base(args[0]) == "cat" {
		fmt.Fprint(tty, "\nPressione Enter para voltar...")
		_, _ = bufio.NewReader(tty).ReadString('\n')
	}
	return nil
}

func firstAvailable(names ...string) []string {
	for _, name := range names {
		if _, err := exec.LookPath(name); err == nil {
			return []string{name}
		}
	}
	return nil
}

// commandWords accepts quotes and escapes, but deliberately does not invoke a shell.
func commandWords(s string) ([]string, error) {
	var words []string
	var b strings.Builder
	var quote rune
	escape, started := false, false
	for _, r := range s {
		if escape {
			b.WriteRune(r)
			escape = false
			started = true
		} else if r == '\\' && quote != '\'' {
			escape, started = true, true
		} else if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				b.WriteRune(r)
			}
		} else if r == '\'' || r == '"' {
			quote, started = r, true
		} else if unicode.IsSpace(r) {
			if started {
				words = append(words, b.String())
				b.Reset()
				started = false
			}
		} else {
			b.WriteRune(r)
			started = true
		}
	}
	if quote != 0 || escape {
		return nil, fmt.Errorf("aspas ou escape incompletos")
	}
	if started {
		words = append(words, b.String())
	}
	return words, nil
}

func (t *externalTool) diff(tty *os.File) error {
	output, err := gitDiff(t.dir, t.path)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp("", "files-tui-diff-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(output); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	viewer := firstAvailable("less", "bat", "cat")
	if len(viewer) == 0 {
		return fmt.Errorf("nenhum visualizador disponível")
	}
	if viewer[0] == "bat" {
		viewer = append(viewer, "--paging=always")
	}
	viewer = append(viewer, "--", file.Name())
	cmd := exec.Command(viewer[0], viewer[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", viewer[0], err)
	}
	if viewer[0] == "cat" {
		fmt.Fprint(tty, "\nPressione Enter para voltar...")
		_, _ = bufio.NewReader(tty).ReadString('\n')
	}
	return nil
}

func gitDiff(dir, path string) ([]byte, error) {
	root, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, fmt.Errorf("Git indisponível ou item fora de um repositório")
	}
	repo := strings.TrimSuffix(string(root), "\n")
	rel, err := filepath.Rel(repo, path)
	if err != nil {
		return nil, err
	}
	var result bytes.Buffer
	for _, section := range []struct {
		title string
		args  []string
	}{
		{"Staged (índice x HEAD)", []string{"--cached"}},
		{"Não staged (arquivos x índice)", nil},
	} {
		args := []string{"-C", repo, "--literal-pathspecs", "diff", "--no-ext-diff", "--no-textconv", "--no-color"}
		args = append(args, section.args...)
		args = append(args, "--", rel)
		cmd := exec.Command("git", args...)
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("erro ao obter diff Git: %w", err)
		}
		fmt.Fprintf(&result, "=== %s ===\n", section.title)
		if len(out) == 0 {
			result.WriteString("(sem alterações)\n")
		} else {
			result.Write(out)
			if out[len(out)-1] != '\n' {
				result.WriteByte('\n')
			}
		}
		result.WriteByte('\n')
	}
	cmd := exec.Command("git", "-C", repo, "--literal-pathspecs", "ls-files", "--others", "--exclude-standard", "-z", "--", rel)
	untracked, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar arquivos não rastreados: %w", err)
	}
	if len(untracked) > 0 {
		result.WriteString("Há arquivos não rastreados; eles não aparecem no diff Git.\n")
	}
	return result.Bytes(), nil
}
