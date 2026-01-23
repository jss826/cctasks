package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
	"github.com/jss826/cctasks/internal/ui"
)

// EditModel handles the task edit/create screen
type EditModel struct {
	task       *data.Task
	taskStore  *data.TaskStore
	groupStore *data.GroupStore
	isNew      bool
	width      int
	height     int

	// Form fields
	subjectInput textinput.Model
	descInput    textarea.Model
	ownerInput   textinput.Model

	// Selectors
	statusIdx int
	groupIdx  int

	// Focus management
	focusIdx int // 0=subject, 1=desc, 2=status, 3=group, 4=owner

	// Available options
	statuses []string
	groups   []string
}

// NewEditModel creates a new EditModel
func NewEditModel(task *data.Task, taskStore *data.TaskStore, groupStore *data.GroupStore, isNew bool) EditModel {
	// Subject input
	subjectInput := textinput.New()
	subjectInput.Placeholder = "Task subject"
	subjectInput.CharLimit = 200
	subjectInput.Width = 50
	subjectInput.Focus()

	// Description input
	descInput := textarea.New()
	descInput.Placeholder = "Task description..."
	descInput.CharLimit = 2000
	descInput.SetWidth(50)
	descInput.SetHeight(5)

	// Owner input
	ownerInput := textinput.New()
	ownerInput.Placeholder = "Owner (optional)"
	ownerInput.CharLimit = 50
	ownerInput.Width = 30

	// Statuses
	statuses := []string{"pending", "in_progress", "completed"}

	// Groups
	groups := append([]string{""}, groupStore.GetGroupNames()...)

	m := EditModel{
		taskStore:    taskStore,
		groupStore:   groupStore,
		isNew:        isNew,
		subjectInput: subjectInput,
		descInput:    descInput,
		ownerInput:   ownerInput,
		statuses:     statuses,
		groups:       groups,
	}

	if isNew {
		// New task defaults
		m.task = &data.Task{
			Status:    "pending",
			Blocks:    []string{},
			BlockedBy: []string{},
		}
		m.statusIdx = 0
		m.groupIdx = 0
	} else {
		// Copy existing task
		taskCopy := *task
		m.task = &taskCopy
		m.subjectInput.SetValue(task.Subject)
		m.descInput.SetValue(task.Description)
		m.ownerInput.SetValue(task.Owner)

		// Find status index
		for i, s := range statuses {
			if s == task.Status {
				m.statusIdx = i
				break
			}
		}

		// Find group index
		taskGroup := data.GetTaskGroup(*task)
		for i, g := range groups {
			if g == taskGroup {
				m.groupIdx = i
				break
			}
		}
	}

	return m
}

// Init initializes the model
func (m EditModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s", "ctrl+enter":
			return m, m.save()
		case "esc":
			return m, func() tea.Msg {
				return CancelEditMsg{}
			}
		case "tab", "shift+tab":
			// Navigate fields
			if msg.String() == "tab" {
				m.focusIdx = (m.focusIdx + 1) % 5
			} else {
				m.focusIdx = (m.focusIdx + 4) % 5
			}
			m.updateFocus()
			return m, nil
		case "up", "down":
			// Handle selector navigation when focused on status or group
			if m.focusIdx == 2 {
				// Status selector
				if msg.String() == "up" && m.statusIdx > 0 {
					m.statusIdx--
				} else if msg.String() == "down" && m.statusIdx < len(m.statuses)-1 {
					m.statusIdx++
				}
				return m, nil
			} else if m.focusIdx == 3 {
				// Group selector
				if msg.String() == "up" && m.groupIdx > 0 {
					m.groupIdx--
				} else if msg.String() == "down" && m.groupIdx < len(m.groups)-1 {
					m.groupIdx++
				}
				return m, nil
			}
		}
	}

	// Update focused input
	switch m.focusIdx {
	case 0:
		m.subjectInput, cmd = m.subjectInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	case 4:
		m.ownerInput, cmd = m.ownerInput.Update(msg)
	}

	return m, cmd
}

func (m *EditModel) updateFocus() {
	m.subjectInput.Blur()
	m.descInput.Blur()
	m.ownerInput.Blur()

	switch m.focusIdx {
	case 0:
		m.subjectInput.Focus()
	case 1:
		m.descInput.Focus()
	case 4:
		m.ownerInput.Focus()
	}
}

func (m *EditModel) save() tea.Cmd {
	// Validate
	subject := strings.TrimSpace(m.subjectInput.Value())
	if subject == "" {
		return nil // Don't save without subject
	}

	// Update task
	m.task.Subject = subject
	m.task.Description = strings.TrimSpace(m.descInput.Value())
	m.task.Status = m.statuses[m.statusIdx]
	m.task.Owner = strings.TrimSpace(m.ownerInput.Value())

	// Set group
	if m.groupIdx > 0 {
		data.SetTaskGroup(m.task, m.groups[m.groupIdx])
	} else {
		data.SetTaskGroup(m.task, "")
	}

	// Save
	if m.isNew {
		m.taskStore.AddTask(*m.task)
	} else {
		m.taskStore.UpdateTask(*m.task)
	}
	m.taskStore.Save()

	return func() tea.Msg {
		return TaskSavedMsg{Store: m.taskStore}
	}
}

// View renders the edit screen
func (m EditModel) View() string {
	var b strings.Builder

	// Header
	title := "Edit Task"
	if m.isNew {
		title = "New Task"
	} else {
		title = fmt.Sprintf("Edit Task #%s", m.task.ID)
	}
	b.WriteString(ui.Header(title, m.width))
	b.WriteString("\n\n")

	// Subject field
	subjectLabel := ui.InputLabelStyle.Render("Subject:")
	if m.focusIdx == 0 {
		subjectLabel = ui.SelectedStyle.Render("Subject:")
	}
	b.WriteString(subjectLabel)
	b.WriteString("\n")
	if m.focusIdx == 0 {
		b.WriteString(ui.FocusedInputStyle.Render(m.subjectInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.subjectInput.View()))
	}
	b.WriteString("\n\n")

	// Description field
	descLabel := ui.InputLabelStyle.Render("Description:")
	if m.focusIdx == 1 {
		descLabel = ui.SelectedStyle.Render("Description:")
	}
	b.WriteString(descLabel)
	b.WriteString("\n")
	if m.focusIdx == 1 {
		b.WriteString(ui.FocusedInputStyle.Render(m.descInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.descInput.View()))
	}
	b.WriteString("\n\n")

	// Status selector
	statusLabel := ui.InputLabelStyle.Render("Status:")
	if m.focusIdx == 2 {
		statusLabel = ui.SelectedStyle.Render("Status:")
	}
	b.WriteString(statusLabel)
	b.WriteString(" ")

	statusText := m.statuses[m.statusIdx]
	statusIcon := ui.StatusIcon(statusText)
	statusStyle := ui.GetStatusStyle(statusText)
	statusDisplay := statusStyle.Render(fmt.Sprintf("%s %s", statusIcon, statusText))

	if m.focusIdx == 2 {
		b.WriteString(ui.FocusedInputStyle.Render(statusDisplay + " ▼"))
	} else {
		b.WriteString(ui.InputStyle.Render(statusDisplay))
	}
	b.WriteString("\n\n")

	// Group selector
	groupLabel := ui.InputLabelStyle.Render("Group:")
	if m.focusIdx == 3 {
		groupLabel = ui.SelectedStyle.Render("Group:")
	}
	b.WriteString(groupLabel)
	b.WriteString(" ")

	groupText := "(none)"
	if m.groupIdx > 0 && m.groupIdx < len(m.groups) {
		groupText = m.groups[m.groupIdx]
	}

	if m.focusIdx == 3 {
		b.WriteString(ui.FocusedInputStyle.Render(groupText + " ▼"))
	} else {
		b.WriteString(ui.InputStyle.Render(groupText))
	}
	b.WriteString("\n\n")

	// Owner field
	ownerLabel := ui.InputLabelStyle.Render("Owner:")
	if m.focusIdx == 4 {
		ownerLabel = ui.SelectedStyle.Render("Owner:")
	}
	b.WriteString(ownerLabel)
	b.WriteString("\n")
	if m.focusIdx == 4 {
		b.WriteString(ui.FocusedInputStyle.Render(m.ownerInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.ownerInput.View()))
	}
	b.WriteString("\n")

	// Footer
	b.WriteString("\n")
	saveKey := "Ctrl+S"
	keys := [][]string{
		{"Tab", "Next Field"},
		{saveKey, "Save"},
		{"Esc", "Cancel"},
	}
	b.WriteString(ui.Footer(keys, m.width))

	return ui.AppStyle.Render(b.String())
}
