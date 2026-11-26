package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

type gitStatus struct {
	added, staged, modified, untracked, ignored bool
}

func (s gitStatus) merge(other gitStatus) gitStatus {
	return gitStatus{s.added || other.added, s.staged || other.staged, s.modified || other.modified, s.untracked || other.untracked, s.ignored || other.ignored}
}

func (s gitStatus) marker() string {
	var b strings.Builder
	if s.added {
		b.WriteByte('A')
	}
	if s.staged {
		b.WriteByte('S')
	}
	if s.modified {
		b.WriteByte('M')
	}
	if s.untracked {
		b.WriteByte('?')
	}
	if s.ignored {
		b.WriteByte('I')
	}
	return b.String()
}

func (m model) statusFor(path string) gitStatus {
	status := m.git[path]
	for parent := filepath.Dir(path); parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
		if m.git[parent].ignored {
			status.ignored = true
			break
		}
	}
	return status
}

// gitStatuses returns statuses for files and all their ancestor directories.
func gitStatuses(dir string) map[string]gitStatus {
	rootOutput, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil
	}
	root := strings.TrimSuffix(string(rootOutput), "\n")
	output, err := exec.Command("git", "-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignored=matching").Output()
	if err != nil {
		return nil
	}
	statuses := make(map[string]gitStatus)
	records := bytes.Split(output, []byte{0})
	for i := 0; i < len(records); i++ {
		record := records[i]
		if len(record) < 4 || record[2] != ' ' {
			continue
		}
		x, y := record[0], record[1]
		status := gitStatus{}
		if x == '!' && y == '!' {
			status.ignored = true
		} else if x == '?' && y == '?' {
			status.untracked = true
		} else {
			status.added = x == 'A'
			status.staged = x != ' ' && x != '?' && x != 'A'
			status.modified = y != ' ' && y != '?'
		}
		path := filepath.Join(root, string(record[3:]))
		for path != root && path != filepath.Dir(path) {
			statuses[path] = statuses[path].merge(status)
			// Ignored status applies to the ignored item, not its containing folders.
			status.ignored = false
			path = filepath.Dir(path)
		}
		// In porcelain -z, a rename/copy includes a second NUL-delimited source path.
		if x == 'R' || x == 'C' || y == 'R' || y == 'C' {
			i++
		}
	}
	return statuses
}

func (m *model) refreshGit() {
	m.git = make(map[string]gitStatus)
	for _, c := range m.columns {
		for path, status := range gitStatuses(c.path) {
			m.git[path] = status
		}
	}
}
