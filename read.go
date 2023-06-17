package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func read() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalln("Error:", err)
	}
}

type noteFile struct {
	name    string
	content string
}

func (f noteFile) time() string {
	t, err := time.Parse("20060102150405", strings.TrimSuffix(f.name, filepath.Ext(f.name)))
	if err != nil {
		log.Fatalln("Error:", err)
	}
	return t.Format(time.RFC3339)
}

func (f noteFile) preview() string {
	c := regexp.MustCompile(`\n+`).ReplaceAllString(strings.TrimSpace(f.content), "...")
	if len(c) > 50 {
		c = fmt.Sprintf("%s...", c[0:50])
	}
	return c
}

type noteDir struct {
	name      string
	noteFiles []noteFile
}

func (d *noteDir) loadNotes() {

	dir, err := os.ReadDir(path.Join(notesDir, d.name))
	if err != nil {
		log.Fatalln("Error:", err)
	}

	var noteFiles []noteFile
	for _, entry := range dir {

		data, err := os.ReadFile(path.Join(notesDir, d.name, entry.Name()))
		if err != nil {
			log.Fatalln("Error:", err)
		}

		f := noteFile{
			name:    entry.Name(),
			content: string(data),
		}

		noteFiles = append(noteFiles, f)
	}

	d.noteFiles = noteFiles
}

type model struct {
	noteDirs []noteDir
	selected *int

	cursorDir    int
	cursorFile   int
	activeCursor string // "dir" or "file"
}

func initialModel() model {

	dir, err := os.ReadDir(notesDir)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	var noteDirs []noteDir

	for _, entry := range dir {
		if entry.IsDir() {
			noteDirs = append(noteDirs, noteDir{name: entry.Name()})
		}
	}

	return model{
		noteDirs:     noteDirs,
		activeCursor: "dir",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			switch m.activeCursor {
			case "dir":
				if m.cursorDir > 0 {
					m.cursorDir--
				}
			case "file":
				if m.cursorFile > 0 {
					m.cursorFile--
				}
			}

		case "down", "j":
			switch m.activeCursor {
			case "dir":
				if m.cursorDir < len(m.noteDirs)-1 {
					m.cursorDir++
				}
			case "file":
				if m.cursorFile < len(m.noteDirs[*m.selected].noteFiles)-1 {
					m.cursorFile++
				}
			}

		case "left", "h":
			m.activeCursor = "dir"

		case "right", "l":
			if m.selected != nil && m.activeCursor != "file" {
				m.cursorFile = 0
				m.activeCursor = "file"
			}

		case "enter", " ":
			switch m.activeCursor {
			case "dir":
				m.selected = &m.cursorDir
				m.noteDirs[*m.selected].loadNotes()
			case "file":
				// no behavior yet
			}

		case "esc":
			m.selected = nil
		}
	}

	return m, nil
}

func (m model) View() string {

	var noteDirs string

	for i, choice := range m.noteDirs {

		cursorDir := " "
		if m.cursorDir == i {
			cursorDir = lipgloss.NewStyle().Faint(m.activeCursor != "dir").Render(">")
		}

		checked := " "
		if m.selected != nil && *m.selected == i {
			checked = "x"
		}

		noteDirs += fmt.Sprintf("%s [%s] %s\n", cursorDir, checked, choice.name)
	}

	var noteFiles string

	if m.selected != nil {

		dir := m.noteDirs[*m.selected]

		var previews []string

		sort.Slice(dir.noteFiles, func(i, j int) bool {
			return dir.noteFiles[i].name > dir.noteFiles[j].name
		})

		for i, f := range dir.noteFiles {
			preview := fmt.Sprintf("%s\n%s", f.time(), f.preview())
			style := lipgloss.NewStyle().BorderLeft(true).PaddingLeft(1)
			if m.activeCursor == "file" && i == m.cursorFile {
				style = style.BorderStyle(lipgloss.Border{Left: ">"})
			} else {
				style = style.BorderStyle(lipgloss.HiddenBorder())
			}
			previews = append(previews, style.Render(preview))
		}
		noteFiles = strings.Join(previews, "\n\n")
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().MarginRight(2).Render(noteDirs),
			noteFiles,
		),
	)
}
