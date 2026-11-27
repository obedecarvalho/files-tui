package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const previewLimit = 256 * 1024

var version = "dev"

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
	git                                        map[string]gitStatus
	focus                                      int
	allowParent                                bool
	reload                                     bool
	showGit                                    bool
	hidden                                     bool
	order, width, height, start, previewOffset int
	editing                                    bool
	showHelp                                   bool
	hidePreview                                bool
	fullPreview                                bool
	helpOffset                                 int
	selected                                   bool
	message                                    string
}

var helpLines = []string{
	"↑/↓ or k/j    Move selection",
	"→, l, Enter   Enter directory",
	"←, h, Backspace   Go to parent directory",
	"g / G         First / last item",
	"[ / ]         Shift columns",
	"Page Up / Page Down   Scroll preview",
	"p             Show / hide preview",
	"P / Esc       Expand / restore preview",
	".             Toggle hidden files",
	"/             Filter names",
	"Enter / Esc   Confirm / clear filter",
	"s             Change sort order",
	"r             Refresh listing and Git status",
	"v / e / d     View / edit / Git diff",
	"Ctrl+g        Toggle Git indicators",
	"?             Open / close shortcuts",
	"Space         Select and exit",
	"q / Ctrl+c    Exit",
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

func (m *model) current() *column { return &m.columns[m.focus] }
func (m *model) selectedPath() string {
	c := m.current()
	if e := c.chosen(); e != nil {
		return filepath.Join(c.path, e.name)
	}
	return c.path
}
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
		if i+1 < len(m.columns) {
			e := c.chosen()
			if e == nil || e.err != nil || e.info == nil || !e.info.IsDir() || filepath.Join(c.path, e.name) != m.columns[i+1].path {
				m.columns = m.columns[:i+1]
				m.focus = min(m.focus, i)
				m.previewOffset = 0
				break
			}
		}
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
	newName := ""
	if e := c.chosen(); e != nil {
		newName = e.name
	}
	if name != newName {
		m.columns = m.columns[:m.focus+1]
	}
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
		m.message = "Not a directory"
		return
	}
	path := filepath.Join(c.path, e.name)
	if m.focus+1 < len(m.columns) && m.columns[m.focus+1].path == path {
		m.focus++
		m.previewOffset = 0
		m.start = -1
		return
	}
	next, err := loadColumn(path, "", m.hidden, m.order, "")
	if err != nil {
		m.message = err.Error()
		return
	}
	m.columns = append(m.columns[:m.focus+1], next)
	m.focus++
	if m.showGit {
		for path, status := range gitStatuses(next.path) {
			if m.git == nil {
				m.git = make(map[string]gitStatus)
			}
			m.git[path] = status
		}
	}
	m.previewOffset = 0
	m.start = -1
}
func (m *model) back() {
	if m.focus > 0 {
		m.focus--
		m.previewOffset = 0
		m.start = -1
		return
	}
	if !m.allowParent {
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
	m.columns = append([]column{next}, m.columns...)
	if m.showGit {
		for path, status := range gitStatuses(next.path) {
			if m.git == nil {
				m.git = make(map[string]gitStatus)
			}
			m.git[path] = status
		}
	}
	m.previewOffset = 0
	m.start = -1
}

type reloadMsg time.Time

func reloadTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return reloadMsg(t) })
}

func (m model) Init() tea.Cmd {
	if m.reload {
		return reloadTick()
	}
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case externalFinished:
		m.refresh()
		if m.showGit {
			m.refreshGit()
		}
		if msg.err != nil {
			m.message = msg.err.Error()
		}
	case reloadMsg:
		if m.reload {
			m.refresh()
			if m.showGit {
				m.refreshGit()
			}
			return m, reloadTick()
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.showHelp {
			switch key {
			case "?", "esc":
				m.showHelp = false
			case "up", "k":
				m.helpOffset = max(0, m.helpOffset-1)
			case "down", "j":
				m.helpOffset = min(max(0, len(helpLines)-max(1, m.height-2)), m.helpOffset+1)
			case "pgup":
				m.helpOffset = max(0, m.helpOffset-max(1, m.height-2))
			case "pgdown":
				m.helpOffset = min(max(0, len(helpLines)-max(1, m.height-2)), m.helpOffset+max(1, m.height-2))
			}
			return m, nil
		}
		if key == "ctrl+g" {
			m.showGit = !m.showGit
			if m.showGit {
				m.refreshGit()
			}
			return m, nil
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
		if m.fullPreview && key == "esc" {
			m.fullPreview = false
			return m, nil
		}
		if m.fullPreview && key == "q" {
			return m, nil
		}
		m.message = ""
		c := m.current()
		oldSelection := ""
		if e := c.chosen(); e != nil {
			oldSelection = e.name
		}
		switch key {
		case "v", "e", "d":
			e := c.chosen()
			if e == nil {
				m.message = "No item selected"
				break
			}
			if e.err != nil || e.info == nil {
				m.message = "Item is inaccessible"
				break
			}
			if key != "d" && !e.info.Mode().IsRegular() {
				m.message = "Select a regular file"
				break
			}
			return m, externalCommand(key, filepath.Join(c.path, e.name), c.path)
		case "?":
			m.showHelp = true
			m.helpOffset = 0
		case " ":
			m.selected = true
			return m, tea.Quit
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
			if m.showGit {
				m.refreshGit()
			}
		case "/":
			m.editing = true
		case "esc":
			m.refilter("")
		case "pgdown":
			lines := preview(c.path, c.chosen())
			m.previewOffset = min(m.previewOffset+3, max(0, len(lines)-max(1, m.height-2)))
		case "pgup":
			m.previewOffset = max(0, m.previewOffset-3)
		case "p":
			if m.fullPreview {
				m.fullPreview = false
				m.hidePreview = true
			} else {
				m.hidePreview = !m.hidePreview
			}
		case "P":
			m.fullPreview = !m.fullPreview
			if m.fullPreview {
				m.hidePreview = false
			}
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
		if key == "up" || key == "k" || key == "down" || key == "j" || key == "g" || key == "G" || key == "." || key == "s" || key == "r" {
			newSelection := ""
			if e := m.current().chosen(); e != nil {
				newSelection = e.name
			}
			if oldSelection != newSelection {
				m.columns = m.columns[:m.focus+1]
			}
		}
	}
	return m, nil
}

func (m model) slots() int     { return max(1, m.width/24) }
func (m model) autoStart() int { return max(0, m.focus+3-m.slots()) }

func humanSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d bytes", size)
	}
	value := float64(size)
	units := []string{"bytes", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	i := 0
	for value >= 1024 && i < len(units)-1 {
		value /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s (%d bytes)", value, units[i], size)
}

func preview(path string, e *entry) []string {
	if e == nil {
		return []string{"(empty)"}
	}
	lines := []string{e.name}
	if e.link {
		lines = append(lines, "Symbolic link")
		if target, err := os.Readlink(filepath.Join(path, e.name)); err == nil {
			lines = append(lines, "Target: "+strings.Map(func(r rune) rune {
				if unicode.IsControl(r) {
					return '�'
				}
				return r
			}, target))
		}
	}
	if e.err != nil {
		return append(lines, "Error: "+e.err.Error())
	}
	if e.info == nil {
		return lines
	}
	typeName := "Special"
	if e.info.IsDir() {
		typeName = "Directory"
	} else if e.info.Mode().IsRegular() {
		typeName = "Regular file"
	}
	lines = append(lines, "Type: "+typeName, "Size: "+humanSize(e.info.Size()), "Modified: "+e.info.ModTime().Format("2006-01-02 15:04"), "Permissions: "+e.info.Mode().String())
	if !e.info.Mode().IsRegular() {
		return lines
	}
	f, err := os.Open(filepath.Join(path, e.name))
	if err != nil {
		return append(lines, "Error: "+err.Error())
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, previewLimit+1))
	if err != nil {
		return append(lines, "Error: "+err.Error())
	}
	truncated := len(data) > previewLimit
	if truncated {
		data = data[:previewLimit]
	}
	if bytes.IndexByte(data, 0) >= 0 {
		mime := http.DetectContentType(data)
		return append(lines, "", "Detected content: "+mime, "(binary file or unsupported text)")
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
		lines = append(lines, "(preview truncated at 256 KiB)")
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
	if m.showHelp {
		return m.helpView()
	}
	if m.fullPreview {
		return m.previewView()
	}
	orders := []string{"name", "size", "date"}
	header := fit(m.current().path+"  | sort: "+orders[m.order], m.width)
	footer := "? shortcuts  Space select  ↑↓/jk move  ←→/hl enter/back  q quit"
	if m.editing {
		footer = "/" + m.current().filter + "_"
	} else if m.message != "" {
		footer = m.message
	} else if m.current().filter != "" {
		footer = "Filter: " + m.current().filter + "  (Esc clears)  " + footer
	}
	if m.height == 1 {
		return header
	}
	bodyHeight := max(0, m.height-2)
	columns := append([]column(nil), m.columns...)
	c := &columns[len(columns)-1]
	var next column
	if e := c.chosen(); e != nil && e.err == nil && e.info != nil && e.info.IsDir() {
		next, _ = loadColumn(filepath.Join(c.path, e.name), "", m.hidden, m.order, "")
	}
	columns = append(columns, next)
	slots := m.slots()
	start := m.start
	if start < 0 {
		start = m.autoStart()
	}
	start = min(start, len(columns)-1)
	count := min(slots, len(columns)-start)
	showPreview := !m.hidePreview && start+count == len(columns) && slots > count
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
		sharedRight := n < count-1 && width >= 3 && colWidth >= 3 && bodyHeight >= 3
		if sharedRight {
			width++
		}
		contentWidth := max(0, width-1)
		boxed := width >= 3 && bodyHeight >= 3
		contentHeight := bodyHeight
		if boxed {
			contentWidth = width - 2
			contentHeight = bodyHeight - 2
		}
		var lines []string
		if idx < len(columns) {
			c := columns[idx]
			offset := c.offset
			if c.selected < offset {
				offset = c.selected
			}
			if c.selected >= offset+contentHeight && contentHeight > 0 {
				offset = c.selected - contentHeight + 1
			}
			if idx < len(m.columns) {
				m.columns[idx].offset = offset
			}
			for row := 0; row < contentHeight; row++ {
				i := offset + row
				text := ""
				if i < len(c.entries) {
					e := c.entries[i]
					var status gitStatus
					if m.showGit {
						status = m.statusFor(filepath.Join(c.path, e.name))
					}
					text = e.name
					if e.link {
						text += "@"
					}
					if e.info != nil && e.info.IsDir() {
						text += "/"
					}
					if marker := status.marker(); marker != "" {
						text = "[" + marker + "] " + text
					}
					if i == c.selected && idx < len(m.columns) {
						text = "> " + text
					} else {
						text = "  " + text
					}
					text = fit(text, contentWidth)
					style := lipgloss.NewStyle()
					if status.modified {
						style = style.Foreground(lipgloss.Color("3"))
					} else if status.added || status.staged {
						style = style.Foreground(lipgloss.Color("2"))
					} else if status.untracked {
						style = style.Foreground(lipgloss.Color("6"))
					} else if status.ignored {
						style = style.Foreground(lipgloss.Color("8"))
					} else if e.info != nil && e.info.IsDir() {
						style = style.Foreground(lipgloss.Color("4"))
					}
					if i == c.selected && idx < len(m.columns) {
						style = style.Reverse(true)
						if idx == m.focus {
							style = style.Foreground(lipgloss.Color("2")).Bold(true)
						}
					}
					text = style.Render(text)
				} else {
					text = fit("", contentWidth)
				}
				lines = append(lines, text)
			}
		} else {
			c := m.current()
			p := preview(c.path, c.chosen())
			offset := min(m.previewOffset, max(0, len(p)-1))
			for row := 0; row < contentHeight; row++ {
				text := ""
				if offset+row < len(p) {
					text = p[offset+row]
				}
				lines = append(lines, fit(text, contentWidth))
			}
		}
		if boxed {
			color := lipgloss.Color("7")
			if idx == m.focus {
				color = lipgloss.Color("2")
			}
			border := lipgloss.NewStyle().Foreground(color)
			leftBorder := border
			if idx-1 == m.focus {
				leftBorder = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
			}
			rightTop, rightSide, rightBottom := "┐", "│", "┘"
			if sharedRight {
				rightTop, rightSide, rightBottom = "", "", ""
			}
			rows[0] += leftBorder.Render("┌") + border.Render(strings.Repeat("─", width-2)+rightTop)
			for row, line := range lines {
				rows[row+1] += leftBorder.Render("│") + line + border.Render(rightSide)
			}
			rows[bodyHeight-1] += leftBorder.Render("└") + border.Render(strings.Repeat("─", width-2)+rightBottom)
			continue
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

func (m model) previewView() string {
	header := fit(m.selectedPath(), m.width)
	if m.height == 1 {
		return header
	}
	lines := preview(m.current().path, m.current().chosen())
	height := max(0, m.height-2)
	offset := min(m.previewOffset, max(0, len(lines)-height))
	rows := []string{header}
	for row := 0; row < height; row++ {
		text := ""
		if offset+row < len(lines) {
			text = lines[offset+row]
		}
		rows = append(rows, fit(text, m.width))
	}
	footer := "P/Esc back  p hide preview  PgUp/PgDown scroll  ? shortcuts"
	if m.editing {
		footer = "/" + m.current().filter + "_"
	} else if m.message != "" {
		footer = m.message
	}
	return strings.Join(append(rows, fit(footer, m.width)), "\n")
}

func (m model) helpView() string {
	rows := []string{fit("Shortcuts", m.width)}
	if m.height == 1 {
		return rows[0]
	}
	bodyHeight := m.height - 2
	offset := min(m.helpOffset, max(0, len(helpLines)-max(1, bodyHeight)))
	for i := 0; i < bodyHeight; i++ {
		line := ""
		if offset+i < len(helpLines) {
			line = helpLines[offset+i]
		}
		rows = append(rows, fit(line, m.width))
	}
	rows = append(rows, fit("↑↓/jk or PgUp/PgDown scroll  ?/Esc close", m.width))
	return strings.Join(rows, "\n")
}

const usage = `fastfile %s
Usage: fastfile [--allow-parent] [--reload] [--git] [--help] [directory]

Argument:
  directory       Starting directory (default: current directory).

Options:
  --allow-parent  Allow navigation above the starting directory.
  --reload        Refresh the listing and active Git statuses every 5 seconds.
  --git           Enable Git status indicators.
  --help          Show this help and the version.
`

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--help" {
			fmt.Printf(usage, version)
			return
		}
	}
	path := "."
	allowParent, reload, showGit, hasPath := false, false, false, false
	for _, arg := range os.Args[1:] {
		if arg == "--allow-parent" {
			allowParent = true
		} else if arg == "--reload" {
			reload = true
		} else if arg == "--git" {
			showGit = true
		} else if !hasPath && !strings.HasPrefix(arg, "-") {
			path, hasPath = arg, true
		} else {
			fmt.Fprintf(os.Stderr, usage, version)
			os.Exit(2)
		}
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
	m := model{columns: []column{c}, start: -1, allowParent: allowParent, reload: reload, showGit: showGit}
	if m.showGit {
		m.refreshGit()
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "terminal unavailable:", err)
		os.Exit(1)
	}
	defer tty.Close()
	final, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithInput(tty), tea.WithOutput(tty)).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if result := final.(model); result.selected {
		fmt.Println(result.selectedPath())
	}
}
