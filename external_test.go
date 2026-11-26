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
	out, err := gitDiff(dir, path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	if !strings.Contains(text, "+staged") || !strings.Contains(text, "+unstaged") {
		t.Fatalf("missing sections: %s", text)
	}
	untracked := filepath.Join(dir, "new")
	if err := os.WriteFile(untracked, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	out, err = gitDiff(dir, untracked)
	if err != nil || !strings.Contains(string(out), "não rastreados") {
		t.Fatalf("untracked: %s, %v", out, err)
	}
	if _, err := gitDiff(t.TempDir(), path); err == nil {
		t.Fatal("expected non-repository error")
	}
}
