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

func (d noteDir) notePreviews() (previews []string) {

	sort.Slice(d.noteFiles, func(i, j int) bool {
		return d.noteFiles[i].name > d.noteFiles[j].name
	})

	for _, f := range d.noteFiles {
		previews = append(previews, fmt.Sprintf("%s\n%s", f.time(), f.preview()))
	}
	return
}

type model struct {
	noteDirs []noteDir
	cursor   int
	selected *int
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
		noteDirs: noteDirs,
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
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.noteDirs)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.selected = &m.cursor
			m.noteDirs[*m.selected].loadNotes()

		case "esc":
			m.selected = nil
		}
	}

	return m, nil
}

func (m model) View() string {

	var noteDirs string

	for i, choice := range m.noteDirs {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if m.selected != nil && *m.selected == i {
			checked = "x"
		}

		noteDirs += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.name)
	}

	var noteFiles string

	if m.selected != nil {

		noteFiles = strings.Join(m.noteDirs[*m.selected].notePreviews(), "\n\n")
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().MarginRight(2).Render(noteDirs),
			noteFiles,
		),
	)
}
