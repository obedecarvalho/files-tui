package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const previewLimit = 256 * 1024

type entry struct {
	name string
	info os.FileInfo
	link bool
	err  error
}

type column struct {
	path             string
	entries          []entry
	selected, offset int
	filter           string
}

type model struct {
	columns                                    []column
	hidden                                     bool
	order, width, height, start, previewOffset int
	editing                                    bool
	message                                    string
}

func loadColumn(path, filter string, hidden bool, order int, selected string) (column, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return column{}, err
	}
	c := column{path: path, filter: filter}
	for _, item := range items {
		name := item.Name()
		if !hidden && strings.HasPrefix(name, ".") {
			continue
		}
		if !strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
			continue
		}
		link := item.Type()&os.ModeSymlink != 0
		info, err := os.Stat(filepath.Join(path, name))
		if err != nil {
			info, _ = item.Info()
		}
		c.entries = append(c.entries, entry{name, info, link, err})
	}
	sort.SliceStable(c.entries, func(i, j int) bool {
		a, b := c.entries[i], c.entries[j]
		ad, bd := a.info != nil && a.info.IsDir(), b.info != nil && b.info.IsDir()
		if ad != bd {
			return ad
		}
		if a.info != nil && b.info != nil {
			if order == 1 && a.info.Size() != b.info.Size() {
				return a.info.Size() > b.info.Size()
			}
			if order == 2 && !a.info.ModTime().Equal(b.info.ModTime()) {
				return a.info.ModTime().After(b.info.ModTime())
			}
		}
		return strings.ToLower(a.name) < strings.ToLower(b.name)
	})
	for i, e := range c.entries {
		if e.name == selected {
			c.selected = i
			break
		}
	}
	return c, nil
}

func (m *model) current() *column { return &m.columns[len(m.columns)-1] }
func (c *column) chosen() *entry {
	if len(c.entries) == 0 {
		return nil
	}
	return &c.entries[c.selected]
}
func (m *model) refresh() {
	for i := range m.columns {
		old := m.columns[i]
		name := ""
		if e := old.chosen(); e != nil {
			name = e.name
		}
		c, err := loadColumn(old.path, old.filter, m.hidden, m.order, name)
		if err != nil {
			m.message = err.Error()
			continue
		}
		m.columns[i] = c
	}
}
func (m *model) refilter(filter string) {
	c := m.current()
	name := ""
	if e := c.chosen(); e != nil {
		name = e.name
	}
	next, err := loadColumn(c.path, filter, m.hidden, m.order, name)
	if err != nil {
		m.message = err.Error()
		return
	}
	*c = next
	m.previewOffset = 0
}
func (m *model) follow() {
	c := m.current()
	e := c.chosen()
	if e == nil {
		return
	}
	if e.err != nil {
		m.message = e.err.Error()
		return
	}
	if e.info == nil || !e.info.IsDir() {
		m.message = "Não é um diretório"
		return
	}
	next, err := loadColumn(filepath.Join(c.path, e.name), "", m.hidden, m.order, "")
	if err != nil {
		m.message = err.Error()
		return
	}
	m.columns = append(m.columns, next)
	m.previewOffset = 0
	m.start = -1
}
func (m *model) back() {
	if len(m.columns) > 1 {
		m.columns = m.columns[:len(m.columns)-1]
		m.refresh()
		m.previewOffset = 0
		m.start = -1
		return
	}
	c := m.current()
	parent := filepath.Dir(c.path)
	if parent == c.path {
		return
	}
	next, err := loadColumn(parent, "", m.hidden, m.order, filepath.Base(c.path))
	if err != nil {
		m.message = err.Error()
		return
	}
	m.columns[0] = next
	m.previewOffset = 0
	m.start = -1
}
func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.editing {
			switch key {
			case "enter":
				m.editing = false
			case "esc":
				m.editing = false
				m.refilter("")
			case "backspace":
				r := []rune(m.current().filter)
				if len(r) > 0 {
					m.refilter(string(r[:len(r)-1]))
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.refilter(m.current().filter + string(msg.Runes))
				}
			}
			return m, nil
		}
		m.message = ""
		c := m.current()
		switch key {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if c.selected > 0 {
				c.selected--
				m.previewOffset = 0
			}
			m.start = -1
		case "down", "j":
			if c.selected+1 < len(c.entries) {
				c.selected++
				m.previewOffset = 0
			}
			m.start = -1
		case "g":
			c.selected = 0
			m.previewOffset = 0
			m.start = -1
		case "G":
			if len(c.entries) > 0 {
				c.selected = len(c.entries) - 1
			}
			m.previewOffset = 0
			m.start = -1
		case "right", "l", "enter":
			m.follow()
		case "left", "h", "backspace":
			m.back()
		case ".":
			m.hidden = !m.hidden
			m.refresh()
		case "s":
			m.order = (m.order + 1) % 3
			m.refresh()
		case "r":
			m.refresh()
		case "/":
			m.editing = true
		case "esc":
			m.refilter("")
		case "pgdown":
			m.previewOffset += max(1, m.height-6)
		case "pgup":
			m.previewOffset = max(0, m.previewOffset-max(1, m.height-6))
		case "[":
			if m.start < 0 {
				m.start = m.autoStart()
			}
			m.start = max(0, m.start-1)
		case "]":
			if m.start < 0 {
				m.start = m.autoStart()
			}
			m.start = min(len(m.columns)-1, m.start+1)
		}
	}
	return m, nil
}

func (m model) slots() int     { return max(1, m.width/24) }
func (m model) autoStart() int { return max(0, len(m.columns)-m.slots()) }

func preview(path string, e *entry) []string {
	if e == nil {
		return []string{"(vazio)"}
	}
	lines := []string{e.name}
	if e.link {
		lines = append(lines, "Link simbólico")
	}
	if e.err != nil {
		return append(lines, "Erro: "+e.err.Error())
	}
	if e.info == nil {
		return lines
	}
	lines = append(lines, fmt.Sprintf("Tamanho: %d bytes", e.info.Size()), "Modificado: "+e.info.ModTime().Format("2006-01-02 15:04"), "Permissões: "+e.info.Mode().String())
	if !e.info.Mode().IsRegular() {
		return lines
	}
	f, err := os.Open(filepath.Join(path, e.name))
	if err != nil {
		return append(lines, "Erro: "+err.Error())
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, previewLimit+1))
	if err != nil {
		return append(lines, "Erro: "+err.Error())
	}
	truncated := len(data) > previewLimit
	if truncated {
		data = data[:previewLimit]
	}
	if bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
		return append(lines, "", "(arquivo binário ou texto não suportado)")
	}
	text := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			return r
		}
		return '�'
	}, string(data))
	lines = append(lines, "")
	lines = append(lines, strings.Split(strings.ReplaceAll(text, "\t", "    "), "\n")...)
	if truncated {
		lines = append(lines, "(prévia truncada em 256 KiB)")
	}
	return lines
}

func fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return runewidth.FillRight(runewidth.Truncate(s, width, "…"), width)
}
func (m model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	orders := []string{"nome", "tamanho", "data"}
	header := fit(m.current().path+"  | ordem: "+orders[m.order], m.width)
	footer := "↑↓/jk mover  ←→/hl entrar/voltar  / filtro  . ocultos  s ordem  q sair"
	if m.editing {
		footer = "/" + m.current().filter + "_"
	} else if m.message != "" {
		footer = m.message
	} else if m.current().filter != "" {
		footer = "Filtro: " + m.current().filter + "  (Esc limpa)  " + footer
	}
	if m.height == 1 {
		return header
	}
	bodyHeight := max(0, m.height-2)
	slots := m.slots()
	start := m.start
	if start < 0 {
		start = m.autoStart()
	}
	start = min(start, len(m.columns)-1)
	count := min(slots, len(m.columns)-start)
	showPreview := start+count == len(m.columns) && slots > count
	if showPreview {
		count++
	}
	colWidth := max(1, m.width/count)
	rows := make([]string, bodyHeight)
	for n := 0; n < count; n++ {
		idx := start + n
		width := colWidth
		if n == count-1 {
			width = m.width - colWidth*(count-1)
		}
		contentWidth := max(0, width-1)
		var lines []string
		if idx < len(m.columns) {
			c := m.columns[idx]
			offset := c.offset
			if c.selected < offset {
				offset = c.selected
			}
			if c.selected >= offset+bodyHeight && bodyHeight > 0 {
				offset = c.selected - bodyHeight + 1
			}
			m.columns[idx].offset = offset
			for row := 0; row < bodyHeight; row++ {
				i := offset + row
				text := ""
				if i < len(c.entries) {
					e := c.entries[i]
					text = e.name
					if e.link {
						text += "@"
					}
					if e.info != nil && e.info.IsDir() {
						text += "/"
					}
					if i == c.selected {
						text = "> " + text
					} else {
						text = "  " + text
					}
					text = fit(text, contentWidth)
					if i == c.selected {
						text = lipgloss.NewStyle().Reverse(true).Render(text)
					} else if e.info != nil && e.info.IsDir() {
						text = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(text)
					}
				} else {
					text = fit("", contentWidth)
				}
				lines = append(lines, text)
			}
		} else {
			c := m.current()
			p := preview(c.path, c.chosen())
			offset := min(m.previewOffset, max(0, len(p)-1))
			for row := 0; row < bodyHeight; row++ {
				text := ""
				if offset+row < len(p) {
					text = p[offset+row]
				}
				lines = append(lines, fit(text, contentWidth))
			}
		}
		for row := range rows {
			rows[row] += lines[row]
			if width > 1 {
				rows[row] += "│"
			}
		}
	}
	return header + "\n" + strings.Join(rows, "\n") + func() string {
		if bodyHeight > 0 {
			return "\n"
		}
		return ""
	}() + fit(footer, m.width)
}

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "Uso: files-tui [diretório]")
		os.Exit(2)
	}
	path := "."
	if len(os.Args) == 2 {
		path = os.Args[1]
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	c, err := loadColumn(abs, "", false, 0, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	m := model{columns: []column{c}, start: -1}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
