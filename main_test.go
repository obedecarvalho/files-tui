package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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
	if err := os.WriteFile(filepath.Join(d, "file"), []byte(strings.Repeat("linha\n", 20)), 0644); err != nil {
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
		t.Fatalf("rolagem para cima: %d, esperado %d", m.previewOffset, end-3)
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
	m.noRoot = true
	m.back()
	if m.current().path != filepath.Dir(d) || len(m.columns) != 2 {
		t.Fatal("--no-root should allow navigation to the parent")
	}
}
