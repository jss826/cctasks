package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
	"github.com/jss826/cctasks/internal/ui"
)

// ProjectsModel handles the project selection screen
type ProjectsModel struct {
	projects []data.Project
	cursor   int
	width    int
	height   int
	err      error
	showHelp bool

	// Double-click detection
	lastClickTime time.Time
	lastClickIdx  int
}

// NewProjectsModel creates a new ProjectsModel
func NewProjectsModel() ProjectsModel {
	return ProjectsModel{}
}

// Init initializes the model and loads projects
func (m ProjectsModel) Init() tea.Cmd {
	return func() tea.Msg {
		projects, err := data.ListProjects()
		return projectsLoadedMsg{projects: projects, err: err}
	}
}

type projectsLoadedMsg struct {
	projects []data.Project
	err      error
}

// Update handles messages
func (m ProjectsModel) Update(msg tea.Msg) (ProjectsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case projectsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.projects = msg.projects
		return m, nil

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			// Calculate which item was clicked
			// Header(2) + empty(2) + Title(1) + Line(1) + empty(2) = 8 lines before list
			// If help is shown, add more lines
			headerLines := 8
			if len(m.projects) == 0 || m.showHelp {
				headerLines += 15 // Approximate help text lines
			}
			clickedIdx := msg.Y - headerLines
			if clickedIdx >= 0 && clickedIdx < len(m.projects) {
				now := time.Now()
				isDoubleClick := clickedIdx == m.lastClickIdx && now.Sub(m.lastClickTime) < 400*time.Millisecond

				if isDoubleClick || clickedIdx == m.cursor {
					// Double-click or click on current cursor: select
					m.cursor = clickedIdx
					m.lastClickTime = now
					m.lastClickIdx = clickedIdx
					return m, func() tea.Msg {
						return SelectProjectMsg{Name: m.projects[m.cursor].Name}
					}
				}
				// Single click on different row: move cursor only
				m.cursor = clickedIdx
				m.lastClickTime = now
				m.lastClickIdx = clickedIdx
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter", "right":
			if len(m.projects) > 0 {
				return m, func() tea.Msg {
					return SelectProjectMsg{Name: m.projects[m.cursor].Name}
				}
			}
		case "q":
			return m, tea.Quit
		case "r":
			return m, m.Init()
		case "?":
			m.showHelp = !m.showHelp
		}
	}

	return m, nil
}

// View renders the project selection screen
func (m ProjectsModel) View() string {
	var b strings.Builder

	// Header with version
	title := fmt.Sprintf("cctasks v%s", AppVersion)
	b.WriteString(ui.Header(title, m.width))
	b.WriteString("\n\n")

	// Title
	b.WriteString(ui.TitleStyle.Render("Projects"))
	b.WriteString("\n")
	b.WriteString(ui.HorizontalLine(m.width))
	b.WriteString("\n\n")

	// Error display
	if m.err != nil {
		b.WriteString(ui.ErrorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	// No projects message or help
	if len(m.projects) == 0 || m.showHelp {
		if len(m.projects) == 0 {
			b.WriteString(ui.MutedStyle.Render("No projects found in ~/.claude/tasks/"))
			b.WriteString("\n\n")
		}

		b.WriteString(ui.SubtitleStyle.Render("Setup Guide"))
		b.WriteString("\n\n")
		b.WriteString("Claude Code v2.1.16+ で Task List 機能を有効にする方法:\n\n")
		b.WriteString("1. プロジェクトの ")
		b.WriteString(ui.KeyStyle.Render(".claude/settings.local.json"))
		b.WriteString(" に以下を追加:\n\n")
		b.WriteString(ui.MutedStyle.Render("   {\n"))
		b.WriteString(ui.MutedStyle.Render("     \"env\": {\n"))
		b.WriteString(ui.MutedStyle.Render("       \"CLAUDE_CODE_TASK_LIST_ID\": \""))
		b.WriteString(ui.SelectedStyle.Render("your-project-name"))
		b.WriteString(ui.MutedStyle.Render("\"\n"))
		b.WriteString(ui.MutedStyle.Render("     }\n"))
		b.WriteString(ui.MutedStyle.Render("   }\n\n"))
		b.WriteString("2. タスクは ")
		b.WriteString(ui.KeyStyle.Render("~/.claude/tasks/your-project-name/"))
		b.WriteString(" に保存されます\n\n")
		b.WriteString(ui.MutedStyle.Render("詳細: "))
		b.WriteString(ui.ValueStyle.Render("https://docs.anthropic.com/en/docs/claude-code/interactive-mode#task-list"))
		b.WriteString("\n")

		if len(m.projects) > 0 {
			b.WriteString("\n")
			b.WriteString(ui.HorizontalLine(m.width))
			b.WriteString("\n\n")
		}
	}

	// Project list
	for i, project := range m.projects {
		cursor := "  "
		style := ui.NormalStyle
		if i == m.cursor {
			cursor = "> "
			style = ui.SelectedStyle
		}

		count := ui.CountBadge(project.TaskCount)
		name := style.Render(project.Name)
		line := fmt.Sprintf("%s%s %s", cursor, name, count)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	keys := [][]string{
		// Navigation
		{"↑↓", "Navigate"},
		{"Enter", "Select"},
		{"?", "Help"},
		// Operations
		{"r", "Refresh"},
		// Exit
		{"q", "Quit"},
	}
	b.WriteString(ui.Footer(keys, m.width))

	return b.String()
}
