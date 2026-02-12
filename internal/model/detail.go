package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
	"github.com/jss826/cctasks/internal/ui"
)

// DetailModel handles the task detail screen
type DetailModel struct {
	task       *data.Task
	taskStore  *data.TaskStore
	groupStore *data.GroupStore
	width      int
	height     int

	// Delete confirmation
	confirmDelete bool

	// Scrolling
	scrollOffset int
}

// NewDetailModel creates a new DetailModel
func NewDetailModel(task *data.Task, taskStore *data.TaskStore, groupStore *data.GroupStore) DetailModel {
	return DetailModel{
		task:       task,
		taskStore:  taskStore,
		groupStore: groupStore,
	}
}

// Init initializes the model
func (m DetailModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	// Delete confirmation mode
	if m.confirmDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				// Delete the task
				m.taskStore.DeleteTask(m.task.ID)
				m.taskStore.Save()
				return m, func() tea.Msg {
					return BackToTasksMsg{}
				}
			case "n", "N", "esc":
				m.confirmDelete = false
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.scrollOffset -= 3
			m.clampScroll()
			return m, nil
		case tea.MouseButtonWheelDown:
			m.scrollOffset += 3
			m.clampScroll()
			return m, nil
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "left":
			return m, func() tea.Msg {
				return BackToTasksMsg{}
			}
		case "j", "down":
			taskID := m.task.ID
			return m, func() tea.Msg {
				return NextTaskMsg{CurrentID: taskID}
			}
		case "k", "up":
			taskID := m.task.ID
			return m, func() tea.Msg {
				return PrevTaskMsg{CurrentID: taskID}
			}
		case "pgdown":
			m.scrollOffset += m.viewportHeight()
			m.clampScroll()
			return m, nil
		case "pgup":
			m.scrollOffset -= m.viewportHeight()
			m.clampScroll()
			return m, nil
		case "home":
			m.scrollOffset = 0
			return m, nil
		case "end":
			m.scrollOffset = m.maxScroll()
			return m, nil
		case "e":
			return m, func() tea.Msg {
				return EditTaskMsg{Task: m.task}
			}
		case "s":
			// Cycle status
			m.cycleStatus()
			return m, nil
		case "d":
			m.confirmDelete = true
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *DetailModel) cycleStatus() {
	statuses := []string{"pending", "in_progress", "completed"}
	for i, s := range statuses {
		if s == m.task.Status {
			m.task.Status = statuses[(i+1)%len(statuses)]
			m.taskStore.UpdateTask(*m.task)
			m.taskStore.Save()
			return
		}
	}
}

// buildBody builds the scrollable body content (everything between header and footer)
func (m DetailModel) buildBody() string {
	var b strings.Builder

	// Delete confirmation dialog
	if m.confirmDelete {
		dialog := ui.Confirm(
			"Delete Task",
			fmt.Sprintf("Are you sure you want to delete task #%s?\n\"%s\"", m.task.ID, m.task.Subject),
			"y", "n",
		)
		b.WriteString(dialog)
		b.WriteString("\n\n")
	}

	// Basic info
	b.WriteString(ui.LabelValue("Subject", m.task.Subject))
	b.WriteString("\n")

	statusBadge := ui.StatusBadge(m.task.Status)
	b.WriteString(ui.LabelStyle.Render("Status:") + " " + statusBadge)
	b.WriteString(ui.MutedStyle.Render("  (s: cycle)"))
	b.WriteString("\n")

	group := data.GetTaskGroup(*m.task)
	if group == "" {
		group = "Uncategorized"
	}
	color := m.groupStore.GetGroupColor(group)
	groupBadge := ui.GroupBadge(group, color)
	b.WriteString(ui.LabelStyle.Render("Group:") + " " + groupBadge)
	b.WriteString("\n")

	if m.task.Owner != "" {
		b.WriteString(ui.LabelValue("Owner", m.task.Owner))
		b.WriteString("\n")
	}

	// Description section
	b.WriteString("\n")
	b.WriteString(ui.HorizontalLine(m.width))
	b.WriteString("\n")
	b.WriteString(ui.MutedStyle.Render("Description:"))
	b.WriteString("\n")

	if m.task.Description != "" {
		desc := ui.WordWrap(m.task.Description, m.width-8)
		b.WriteString(desc)
	} else {
		b.WriteString(ui.MutedStyle.Render("(no description)"))
	}
	b.WriteString("\n")

	// Dependencies section
	b.WriteString("\n")
	b.WriteString(ui.HorizontalLine(m.width))
	b.WriteString("\n")
	b.WriteString(ui.MutedStyle.Render("Dependencies:"))
	b.WriteString("\n")

	// Blocks
	b.WriteString("  Blocks:    ")
	if len(m.task.Blocks) > 0 {
		var blockStrs []string
		for _, id := range m.task.Blocks {
			task := m.taskStore.GetTask(id)
			if task != nil {
				blockStrs = append(blockStrs, fmt.Sprintf("#%s %s", id, task.Subject))
			} else {
				blockStrs = append(blockStrs, fmt.Sprintf("#%s", id))
			}
		}
		b.WriteString(strings.Join(blockStrs, ", "))
	} else {
		b.WriteString(ui.MutedStyle.Render("(none)"))
	}
	b.WriteString("\n")

	// BlockedBy
	b.WriteString("  BlockedBy: ")
	if len(m.task.BlockedBy) > 0 {
		var blockedByStrs []string
		for _, id := range m.task.BlockedBy {
			task := m.taskStore.GetTask(id)
			if task != nil {
				blockedByStrs = append(blockedByStrs, fmt.Sprintf("#%s %s", id, task.Subject))
			} else {
				blockedByStrs = append(blockedByStrs, fmt.Sprintf("#%s", id))
			}
		}
		b.WriteString(strings.Join(blockedByStrs, ", "))
	} else {
		b.WriteString(ui.MutedStyle.Render("(none)"))
	}

	return b.String()
}

// viewportHeight returns the number of lines available for body content
func (m DetailModel) viewportHeight() int {
	// header: 2 lines (title + horizontal line) + 1 empty line = 3
	// footer: 1 horizontal line + 1-2 hint lines = 2-3
	// scroll indicators: up to 2 lines
	overhead := 8
	vh := m.height - overhead
	if vh < 5 {
		vh = 5
	}
	return vh
}

// maxScroll returns the maximum valid scroll offset
func (m DetailModel) maxScroll() int {
	body := m.buildBody()
	lines := strings.Split(body, "\n")
	vh := m.viewportHeight()
	if len(lines) <= vh {
		return 0
	}
	return len(lines) - vh
}

// clampScroll ensures scrollOffset is within valid bounds
func (m *DetailModel) clampScroll() {
	max := m.maxScroll()
	if m.scrollOffset > max {
		m.scrollOffset = max
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// View renders the task detail screen
func (m DetailModel) View() string {
	var result strings.Builder

	// Header
	title := fmt.Sprintf("Task #%s", m.task.ID)
	result.WriteString(ui.Header(title, m.width))
	result.WriteString("\n\n")

	// Build body content
	body := m.buildBody()
	bodyLines := strings.Split(body, "\n")
	totalLines := len(bodyLines)
	vh := m.viewportHeight()

	// Clamp scroll offset for display
	scrollOffset := m.scrollOffset
	maxOff := 0
	if totalLines > vh {
		maxOff = totalLines - vh
	}
	if scrollOffset > maxOff {
		scrollOffset = maxOff
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	needsScroll := totalLines > vh

	if !needsScroll {
		// Everything fits, no scrolling needed
		result.WriteString(body)
		result.WriteString("\n")
	} else {
		// Top scroll indicator
		if scrollOffset > 0 {
			result.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  ↑ %d lines above", scrollOffset)))
			result.WriteString("\n")
		}

		// Visible slice
		endIdx := scrollOffset + vh
		if scrollOffset > 0 {
			endIdx-- // account for top indicator line
		}
		remaining := totalLines - endIdx
		if remaining > 0 {
			endIdx-- // account for bottom indicator line
		}
		if endIdx > totalLines {
			endIdx = totalLines
		}

		visibleLines := bodyLines[scrollOffset:endIdx]
		result.WriteString(strings.Join(visibleLines, "\n"))
		result.WriteString("\n")

		// Bottom scroll indicator
		remaining = totalLines - endIdx
		if remaining > 0 {
			result.WriteString(ui.MutedStyle.Render(fmt.Sprintf("  ↓ %d lines below", remaining)))
			result.WriteString("\n")
		}
	}

	// Footer - context-aware
	if m.confirmDelete {
		hints := []ui.KeyHint{
			{Key: "y", Desc: "Confirm", Enabled: true},
			{Key: "n", Desc: "Cancel", Enabled: true},
		}
		result.WriteString(ui.FooterWithHints(hints, m.width))
	} else {
		hints := []ui.KeyHint{
			// Navigation
			{Key: "j/k", Desc: "Next/Prev", Enabled: true},
			{Key: "Esc", Desc: "Back", Enabled: true},
			// Task operations
			{Key: "e", Desc: "Edit", Enabled: true},
			{Key: "s", Desc: "Status", Enabled: true},
			{Key: "d", Desc: "Delete", Enabled: true},
		}
		if needsScroll {
			hints = append(hints, ui.KeyHint{Key: "PgUp/Dn", Desc: "Scroll", Enabled: true})
		}
		hints = append(hints, ui.KeyHint{Key: "q", Desc: "Quit", Enabled: true})
		result.WriteString(ui.FooterWithHints(hints, m.width))
	}

	return result.String()
}
