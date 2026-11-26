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
