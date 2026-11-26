package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestGitStatuses(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	d := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", d}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	git("init", "-q")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	if err := os.Mkdir(filepath.Join(d, "dir"), 0755); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(d, "dir", "name with\nnewline")
	if err := os.WriteFile(name, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-qm", "initial")
	if err := os.WriteFile(name, []byte("staged"), 0644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	if err := os.WriteFile(name, []byte("modified again"), 0644); err != nil {
		t.Fatal(err)
	}
	untracked := filepath.Join(d, "dir", "untracked")
	if err := os.WriteFile(untracked, nil, 0644); err != nil {
		t.Fatal(err)
	}
	statuses := gitStatuses(d)
	if got := statuses[name].marker(); got != "SM" {
		t.Fatalf("file status: %q", got)
	}
	if got := statuses[untracked].marker(); got != "?" {
		t.Fatalf("untracked status: %q", got)
	}
	if got := statuses[filepath.Join(d, "dir")].marker(); got != "SM?" {
		t.Fatalf("directory status: %q", got)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, start: -1, width: 100, height: 10}
	m.refreshGit()
	if strings.Contains(m.View(), "[SM?]") {
		t.Fatal("Git markers should be hidden by default")
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	m = updated.(model)
	if !strings.Contains(m.View(), "[SM?] dir/") {
		t.Fatal("directory marker missing from view")
	}
	git("add", ".")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = updated.(model)
	if got := m.git[filepath.Join(d, "dir")].marker(); got != "AS" {
		t.Fatalf("refresh status: %q", got)
	}
	if got := m.git[untracked].marker(); got != "A" {
		t.Fatalf("new staged file status: %q", got)
	}
	if err := os.WriteFile(untracked, []byte("changed after staging"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := gitStatuses(d)[untracked].marker(); got != "AM" {
		t.Fatalf("new staged file with unstaged changes: %q", got)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	m = updated.(model)
	if m.showGit || strings.Contains(m.View(), "[AS]") {
		t.Fatal("Ctrl+g should hide Git markers")
	}
}

func TestGitStatusesOutsideRepository(t *testing.T) {
	if got := gitStatuses(t.TempDir()); len(got) != 0 {
		t.Fatalf("unexpected statuses: %v", got)
	}
}

func TestGitIgnoredItems(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	d := t.TempDir()
	cmd := exec.Command("git", "-C", d, "init", "-q")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	if err := os.WriteFile(filepath.Join(d, ".gitignore"), []byte("ignored/\n*.log\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ignoredDir := filepath.Join(d, "ignored")
	if err := os.MkdirAll(filepath.Join(ignoredDir, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(ignoredDir, "nested", "file")
	for _, name := range []string{inside, filepath.Join(d, "debug.log"), filepath.Join(d, "new")} {
		if err := os.WriteFile(name, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	m := model{git: gitStatuses(d)}
	for _, path := range []string{ignoredDir, inside, filepath.Join(d, "debug.log")} {
		if got := m.statusFor(path).marker(); got != "I" {
			t.Errorf("%s: got %q, want I", path, got)
		}
	}
	if got := m.statusFor(filepath.Join(d, "new")).marker(); got != "?" {
		t.Errorf("new file: got %q, want ?", got)
	}
}
