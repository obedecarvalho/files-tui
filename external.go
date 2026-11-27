package main

import (
	"bufio"
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
	args, err := configuredCommand(t.action)
	if err != nil {
		return err
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("terminal unavailable: %w", err)
	}
	defer tty.Close()
	if t.action == "d" {
		return t.diff(tty, args)
	}
	args = append(args, "--", t.path)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", args[0], err)
	}
	if t.action == "v" && filepath.Base(args[0]) == "cat" {
		fmt.Fprint(tty, "\nPress Enter to return...")
		_, _ = bufio.NewReader(tty).ReadString('\n')
	}
	return nil
}

func configuredCommand(action string) ([]string, error) {
	key, variable := "pager", "PAGER"
	if action == "e" {
		key, variable = "editor", "EDITOR"
	} else if action == "d" {
		key, variable = "git_diff", ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home directory unavailable: %w", err)
	}
	path := filepath.Join(home, ".config", "fastfile.conf")
	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration: %w", err)
	}
	value := ""
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for line := 1; scanner.Scan(); line++ {
			text := strings.TrimSpace(scanner.Text())
			if text == "" || strings.HasPrefix(text, "#") {
				continue
			}
			name, command, ok := strings.Cut(text, "=")
			if !ok {
				return nil, fmt.Errorf("invalid configuration on line %d", line)
			}
			if strings.TrimSpace(name) == key {
				value = strings.TrimSpace(command)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("configuration: %w", err)
		}
	}
	if value == "" && variable != "" {
		value = os.Getenv(variable)
	}
	if strings.TrimSpace(value) == "" {
		if variable != "" {
			return nil, fmt.Errorf("%s is not set", variable)
		}
		return nil, fmt.Errorf("git_diff is not set in %s", path)
	}
	args, err := commandWords(value)
	if err != nil || len(args) == 0 || args[0] == "" {
		return nil, fmt.Errorf("invalid %s command", key)
	}
	if action == "d" && len(args) != 1 {
		return nil, fmt.Errorf("git_diff must contain only the Git tool name")
	}
	return args, nil
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
		return nil, fmt.Errorf("incomplete quote or escape")
	}
	if started {
		words = append(words, b.String())
	}
	return words, nil
}

func (t *externalTool) diff(tty *os.File, tool []string) error {
	root, err := exec.Command("git", "-C", t.dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("Git unavailable or item outside a repository")
	}
	repo := strings.TrimSuffix(string(root), "\n")
	rel, err := filepath.Rel(repo, t.path)
	if err != nil {
		return err
	}
	for _, cached := range []bool{true, false} {
		args := []string{"-C", repo, "--literal-pathspecs", "difftool", "--no-prompt", "--trust-exit-code", "--tool=" + tool[0]}
		if cached {
			args = append(args, "--cached")
		}
		args = append(args, "--", rel)
		cmd := exec.Command("git", args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git difftool: %w", err)
		}
	}
	return nil
}
