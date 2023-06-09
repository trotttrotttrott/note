package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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
	path    string
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

type editorFinishedMsg struct{ err error }

func (f noteFile) edit() tea.Cmd {

	editor := os.Getenv("EDITOR")

	if editor == "" {
		editor = "vim"
	}

	c := exec.Command(editor, f.path)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
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

		notePath := path.Join(notesDir, d.name, entry.Name())

		data, err := os.ReadFile(notePath)
		if err != nil {
			log.Fatalln("Error:", err)
		}

		f := noteFile{
			name:    entry.Name(),
			content: string(data),
			path:    notePath,
		}

		noteFiles = append(noteFiles, f)
	}

	d.noteFiles = noteFiles
}

type model struct {
	noteDirs []noteDir

	cursorDir   int
	selectedDir *int

	cursorFile   int
	selectedFile *int

	activeCursor string // "dir" or "file"

	err error
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
				if m.cursorFile < len(m.noteDirs[*m.selectedDir].noteFiles)-1 {
					m.cursorFile++
				}
			}

		case "left", "h":
			m.activeCursor = "dir"

		case "right", "l":
			if m.selectedDir != nil && m.activeCursor != "file" {
				m.cursorFile = 0
				m.activeCursor = "file"
			}

		case "enter", " ":
			switch m.activeCursor {
			case "dir":
				m.selectedFile = nil
				m.selectedDir = &m.cursorDir
				m.noteDirs[*m.selectedDir].loadNotes()
			case "file":
				m.selectedFile = &m.cursorFile
			}

		case "e":
			switch m.activeCursor {
			case "file":
				return m, m.noteDirs[*m.selectedDir].noteFiles[*&m.cursorFile].edit()
			}

		case "esc":
			switch m.activeCursor {
			case "dir":
				m.selectedDir = nil
			case "file":
				m.selectedFile = nil
			}
		}

	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.noteDirs[*m.selectedDir].loadNotes()
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
		if m.selectedDir != nil && *m.selectedDir == i {
			checked = "x"
		}

		noteDirs += fmt.Sprintf("%s [%s] %s\n", cursorDir, checked, choice.name)
	}

	var noteFiles []string

	if m.selectedDir != nil {

		dir := m.noteDirs[*m.selectedDir]

		sort.Slice(dir.noteFiles, func(i, j int) bool {
			return dir.noteFiles[i].name > dir.noteFiles[j].name
		})

		for i, f := range dir.noteFiles {

			var body string
			if m.selectedFile != nil && i == *m.selectedFile {
				body = strings.TrimSpace(f.content)
			} else {
				body = f.preview()
			}

			style := lipgloss.NewStyle().BorderLeft(true).PaddingLeft(1)

			if m.activeCursor == "file" && i == m.cursorFile {
				style = style.BorderStyle(lipgloss.Border{Left: ">"})
			} else {
				style = style.BorderStyle(lipgloss.HiddenBorder())
			}
			noteFiles = append(noteFiles, style.Render(
				fmt.Sprintf("%s\n%s", lipgloss.NewStyle().Faint(true).Render(f.time()), body),
			))
		}
	}

	var errMsg string
	if m.err != nil {
		errMsg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("Error: " + m.err.Error() + "\n")
		m.err = nil
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		errMsg,
		lipgloss.NewStyle().Margin(0, 1, 1, 2).Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				lipgloss.NewStyle().MarginRight(2).Render(noteDirs),
				strings.Join(noteFiles, "\n\n"),
			),
		),
	)
}
