package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCommandWords(t *testing.T) {
	got, err := commandWords(`editor --wait "a b" 'c d'`)
	if err != nil || !reflect.DeepEqual(got, []string{"editor", "--wait", "a b", "c d"}) {
		t.Fatalf("words = %q, %v", got, err)
	}
	if _, err := commandWords(`editor "unterminated`); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfiguredCommand(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("EDITOR", "vim")
	t.Setenv("PAGER", "less")
	check := func(action string, want []string) {
		t.Helper()
		got, err := configuredCommand(action)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: got %q, %v; want %q", action, got, err, want)
		}
	}
	check("e", []string{"vim"})
	check("v", []string{"less"})
	if _, err := configuredCommand("d"); err == nil || !strings.Contains(err.Error(), "git_diff is not set") {
		t.Fatalf("expected missing diff configuration: %v", err)
	}
	if err := os.Mkdir(filepath.Join(home, ".config"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".config", "fastfile.conf")
	if err := os.WriteFile(path, []byte("# ferramentas\neditor = code --wait\npager = bat --paging=always\ngit_diff = meld\n"), 0600); err != nil {
		t.Fatal(err)
	}
	check("e", []string{"code", "--wait"})
	check("v", []string{"bat", "--paging=always"})
	check("d", []string{"meld"})
	if err := os.WriteFile(path, []byte("editor = \"incompleto\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := configuredCommand("e"); err == nil {
		t.Fatal("expected invalid command error")
	}
}

func TestGitDiffSections(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	path := filepath.Join(dir, "a [file].txt")
	write := func(text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("original\n")
	run("add", "--", filepath.Base(path))
	run("commit", "-qm", "initial")
	write("staged\n")
	run("add", "--", filepath.Base(path))
	write("unstaged\n")
	log := filepath.Join(t.TempDir(), "comparisons")
	t.Setenv("DIFF_TEST_LOG", log)
	run("config", "difftool.test.cmd", `printf '%s|' "$(cat "$LOCAL")" "$(cat "$REMOTE")" >> "$DIFF_TEST_LOG"`)
	tty, err := os.CreateTemp(t.TempDir(), "tty")
	if err != nil {
		t.Fatal(err)
	}
	defer tty.Close()
	tool := externalTool{dir: dir, path: path}
	if err := tool.diff(tty, []string{"test"}); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(log)
	if err != nil || string(output) != "original|staged|staged|unstaged|" {
		t.Fatalf("comparisons: %q, %v", output, err)
	}
	if err := (&externalTool{dir: t.TempDir(), path: path}).diff(tty, []string{"test"}); err == nil {
		t.Fatal("expected non-repository error")
	}
}
