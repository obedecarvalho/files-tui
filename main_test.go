package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelp(t *testing.T) {
	if os.Getenv("FASTFILE_TEST_HELP") == "1" {
		os.Args = []string{"fastfile", "--help"}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelp$")
	cmd.Env = append(os.Environ(), "FASTFILE_TEST_HELP=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, output)
	}
	for _, want := range []string{
		"fastfile " + version,
		"Usage: fastfile",
		"directory       Starting directory (default: current directory).",
		"--allow-parent  Allow navigation above the starting directory.",
		"--reload        Refresh the listing and active Git statuses every 5 seconds.",
		"--git           Enable Git status indicators.",
		"--help          Show this help and the version.",
	} {
		if !strings.Contains(string(output), want) {
			t.Errorf("help missing %q: %s", want, output)
		}
	}
}

func TestInvalidArgumentShowsHelp(t *testing.T) {
	if os.Getenv("FASTFILE_TEST_INVALID_ARGUMENT") == "1" {
		os.Args = []string{"fastfile", "--invalid"}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestInvalidArgumentShowsHelp$")
	cmd.Env = append(os.Environ(), "FASTFILE_TEST_INVALID_ARGUMENT=1")
	output, err := cmd.Output()
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got %v", err)
	}
	if len(output) != 0 || string(exitErr.Stderr) != fmt.Sprintf(usage, version) {
		t.Fatalf("unexpected output: stdout %q, stderr %q", output, exitErr.Stderr)
	}
}

func TestShortcutHelp(t *testing.T) {
	c, err := loadColumn(t.TempDir(), "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, width: 80, height: 6}
	press := func(key tea.KeyMsg) {
		updated, _ := m.Update(key)
		m = updated.(model)
	}
	question := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	press(question)
	if !m.showHelp || !strings.Contains(m.View(), "Shortcuts") {
		t.Fatal("shortcut help did not open")
	}
	for i := 0; i < len(helpLines); i++ {
		press(tea.KeyMsg{Type: tea.KeyDown})
	}
	if !strings.Contains(m.View(), "Ctrl+c") {
		t.Fatal("last shortcut is not reachable")
	}
	press(tea.KeyMsg{Type: tea.KeyEsc})
	if m.showHelp || m.focus != 0 {
		t.Fatal("escape did not close help without navigating")
	}
	press(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	press(question)
	if m.showHelp || m.current().filter != "?" {
		t.Fatal("question mark should remain available in the filter")
	}
}

func TestLoadFilterSortAndLinks(t *testing.T) {
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "Dir"), 0755); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"z.txt": "1234567890", "Alpha.txt": "1", ".hidden": "x"} {
		if err := os.WriteFile(filepath.Join(d, name), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("missing", filepath.Join(d, "broken")); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.entries) != 4 || c.entries[0].name != "Dir" || c.entries[1].name != "z.txt" {
		t.Fatalf("unexpected entries: %+v", c.entries)
	}
	c, err = loadColumn(d, "ALPHA", true, 0, "")
	if err != nil || len(c.entries) != 1 || c.entries[0].name != "Alpha.txt" {
		t.Fatalf("filter: %+v %v", c.entries, err)
	}
	c, err = loadColumn(d, "broken", true, 0, "")
	if err != nil || !c.entries[0].link || c.entries[0].err == nil {
		t.Fatalf("broken link: %+v %v", c.entries, err)
	}
}

func TestNavigationFilterAndPreview(t *testing.T) {
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "child"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "file"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, start: -1, width: 80, height: 20}
	m.follow()
	if len(m.columns) != 2 {
		t.Fatal("did not enter")
	}
	m.back()
	if m.current().path != d {
		t.Fatal("did not go back")
	}
	m.refilter("FILE")
	if len(m.current().entries) != 1 {
		t.Fatal("filter failed")
	}
	if got := strings.Join(preview(d, m.current().chosen()), "\n"); !strings.Contains(got, "hello") {
		t.Fatal(got)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(model)
	if result.current().filter != "" {
		t.Fatal("escape did not clear filter")
	}
	if got := m.View(); !strings.Contains(got, "file") {
		t.Fatal("missing entry in view")
	}
}

func TestPreviewScrollIsBounded(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "file"), []byte(strings.Repeat("line\n", 20)), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, height: 10, width: 80, start: -1}
	for i := 0; i < 100; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
		m = updated.(model)
	}
	end := m.previewOffset
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = updated.(model)
	if m.previewOffset != end-3 {
		t.Fatalf("scroll up: %d, expected %d", m.previewOffset, end-3)
	}
}

func TestPreviewMetadataAndBinary(t *testing.T) {
	d := t.TempDir()
	data := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 2048)...)
	if err := os.WriteFile(filepath.Join(d, "image"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("image", filepath.Join(d, "link")); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		want []string
	}{
		{"image", []string{"Type: Regular file", "Size: 2.0 KiB (2056 bytes)", "Detected content: image/png", "binary file"}},
		{"link", []string{"Symbolic link", "Target: image", "image/png"}},
	} {
		c, err := loadColumn(d, tc.name, false, 0, tc.name)
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Join(preview(d, c.chosen()), "\n")
		for _, want := range tc.want {
			if !strings.Contains(got, want) {
				t.Errorf("%s: missing %q in %s", tc.name, want, got)
			}
		}
	}
}

func TestPreviewInvalidUTF8(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		want string
	}{
		{"invalid", []byte("hello \xffworld"), "hello �world"},
		{"truncated", append([]byte(strings.Repeat("a", previewLimit-1)), []byte("é")...), strings.Repeat("a", previewLimit-1) + "�\n(preview truncated at 256 KiB)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := t.TempDir()
			if err := os.WriteFile(filepath.Join(d, "file"), tc.data, 0644); err != nil {
				t.Fatal(err)
			}
			c, err := loadColumn(d, "", false, 0, "")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.Join(preview(d, c.chosen()), "\n")
			if !strings.Contains(got, tc.want) {
				t.Fatal("preview did not contain the expected replacement text")
			}
		})
	}
}

func TestTogglePreview(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "file"), []byte("unique preview content"), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, start: -1, width: 120, height: 20}
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")}
	if !strings.Contains(m.View(), "unique preview content") {
		t.Fatal("preview should be enabled by default")
	}
	updated, _ := m.Update(key)
	m = updated.(model)
	if strings.Contains(m.View(), "unique preview content") {
		t.Fatal("preview should be hidden")
	}
	updated, _ = m.Update(key)
	m = updated.(model)
	if !strings.Contains(m.View(), "unique preview content") {
		t.Fatal("preview should be visible again")
	}
	m.editing = true
	updated, _ = m.Update(key)
	m = updated.(model)
	if m.hidePreview || m.current().filter != "p" {
		t.Fatal("p should remain usable in filters")
	}
}

func TestFullPreview(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "file"), []byte("preview content"), 0644); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, start: -1, width: 40, height: 12, hidePreview: true}
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("P")}
	updated, _ := m.Update(key)
	m = updated.(model)
	if !m.fullPreview || m.hidePreview || !strings.Contains(m.View(), "preview content") || strings.Contains(m.View(), "┌") {
		t.Fatal("P should display only the expanded preview")
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(model)
	if cmd != nil || !m.fullPreview {
		t.Fatal("q should not quit while the expanded preview is open")
	}
	updated, _ = m.Update(key)
	m = updated.(model)
	if m.fullPreview || !strings.Contains(m.View(), "┌") {
		t.Fatal("P should restore the normal layout")
	}
	updated, _ = m.Update(key)
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.fullPreview {
		t.Fatal("escape should close the expanded preview")
	}
	updated, _ = m.Update(key)
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = updated.(model)
	if m.fullPreview || !m.hidePreview {
		t.Fatal("p should hide the expanded preview")
	}
}

func TestLookaheadAndPreservedColumns(t *testing.T) {
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "child"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "child", "inside"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "other"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, start: -1, width: 120, height: 12}
	if !strings.Contains(m.View(), "inside") || m.focus != 0 {
		t.Fatal("lookahead should show the child without changing focus")
	}
	m.follow()
	before := strings.Count(strings.Split(m.View(), "\n")[1], "┌")
	m.back()
	if len(m.columns) != 2 || m.focus != 0 || !strings.Contains(m.View(), "inside") {
		t.Fatal("back should preserve the child column")
	}
	if after := strings.Count(strings.Split(m.View(), "\n")[1], "┌"); after != before {
		t.Fatalf("back changed the number of visible columns: %d to %d", before, after)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	if len(m.columns) != 1 || strings.Contains(m.View(), "inside") {
		t.Fatal("changing selection should discard stale child columns")
	}
}

func TestInitialDirectoryBoundary(t *testing.T) {
	d := t.TempDir()
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, start: -1}
	m.back()
	if m.current().path != d || len(m.columns) != 1 {
		t.Fatal("default navigation escaped the initial directory")
	}
	m.allowParent = true
	m.back()
	if m.current().path != filepath.Dir(d) || len(m.columns) != 2 {
		t.Fatal("--allow-parent should allow navigation to the parent")
	}
}

func TestSelectedPath(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "file"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}}
	if got := m.selectedPath(); got != filepath.Join(d, "file") {
		t.Fatalf("selected path: %q", got)
	}
	m.refilter("no match")
	if got := m.selectedPath(); got != d {
		t.Fatalf("empty selection path: %q", got)
	}
}

func TestSpaceSelectsAndQuitDoesNot(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "file"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	selected := updated.(model)
	if !selected.selected || cmd == nil || selected.selectedPath() != filepath.Join(d, "file") {
		t.Fatal("space should select the focused path and quit")
	}
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if updated.(model).selected || cmd == nil {
		t.Fatal("q should quit without selecting")
	}
	m.editing = true
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	filtered := updated.(model)
	if filtered.selected || cmd != nil || filtered.current().filter != " " {
		t.Fatal("space should be usable in a filter")
	}
}

func TestReload(t *testing.T) {
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "child"), 0755); err != nil {
		t.Fatal(err)
	}
	c, err := loadColumn(d, "", false, 0, "child")
	if err != nil {
		t.Fatal(err)
	}
	m := model{columns: []column{c}, reload: true}
	if m.Init() == nil || (model{}).Init() != nil {
		t.Fatal("reload timer should only start when enabled")
	}
	m.follow()
	if err := os.WriteFile(filepath.Join(d, "child", "new"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	updated, cmd := m.Update(reloadMsg{})
	m = updated.(model)
	if cmd == nil || len(m.current().entries) != 1 || m.current().entries[0].name != "new" {
		t.Fatal("reload did not update the current column and schedule another tick")
	}
	if err := os.Remove(filepath.Join(d, "child", "new")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(d, "child")); err != nil {
		t.Fatal(err)
	}
	updated, _ = m.Update(reloadMsg{})
	m = updated.(model)
	if len(m.columns) != 1 || m.focus != 0 {
		t.Fatal("reload should discard columns for a removed directory")
	}
}
