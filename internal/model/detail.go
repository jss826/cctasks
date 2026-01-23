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
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return BackToTasksMsg{}
			}
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

// View renders the task detail screen
func (m DetailModel) View() string {
	var b strings.Builder

	// Header
	title := fmt.Sprintf("Task #%s", m.task.ID)
	b.WriteString(ui.Header(title, m.width))
	b.WriteString("\n\n")

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
	b.WriteString(ui.HorizontalLine(m.width - 4))
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
	b.WriteString(ui.HorizontalLine(m.width - 4))
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
	b.WriteString("\n")

	// Footer
	b.WriteString("\n")
	keys := [][]string{
		{"Esc", "Back"},
		{"e", "Edit"},
		{"s", "Status"},
		{"d", "Delete"},
		{"q", "Quit"},
	}
	b.WriteString(ui.Footer(keys, m.width))

	return ui.AppStyle.Render(b.String())
}
