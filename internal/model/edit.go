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
	subjectInput   textinput.Model
	descInput      textarea.Model
	ownerInput     textinput.Model
	blocksInput    textinput.Model
	blockedByInput textinput.Model

	// Selectors
	statusIdx int
	groupIdx  int

	// Focus management
	focusIdx int // 0=subject, 1=desc, 2=status, 3=group, 4=owner, 5=blocks, 6=blockedBy

	// Available options
	statuses []string
	groups   []string

	// Task picker mode (for blocks/blockedBy)
	pickerActive   bool
	pickerForField int // 5=blocks, 6=blockedBy
	pickerSearch   textinput.Model
	pickerTasks    []data.Task // filtered tasks
	pickerCursor   int
	pickerSelected map[string]bool // selected task IDs
}

// NewEditModel creates a new EditModel
func NewEditModel(task *data.Task, taskStore *data.TaskStore, groupStore *data.GroupStore, isNew bool) EditModel {
	// Subject input
	subjectInput := textinput.New()
	subjectInput.Placeholder = "Task subject"
	subjectInput.CharLimit = 200
	subjectInput.Width = 60
	subjectInput.Prompt = "> "
	subjectInput.Focus()

	// Description input
	descInput := textarea.New()
	descInput.Placeholder = "Task description..."
	descInput.CharLimit = 2000
	descInput.SetWidth(60)
	descInput.SetHeight(4)
	descInput.ShowLineNumbers = false
	descInput.Prompt = "  "

	// Owner input
	ownerInput := textinput.New()
	ownerInput.Placeholder = "Owner (optional)"
	ownerInput.CharLimit = 50
	ownerInput.Width = 40
	ownerInput.Prompt = "> "

	// Blocks input
	blocksInput := textinput.New()
	blocksInput.Placeholder = "Task IDs (comma-separated, e.g. 1,2,3)"
	blocksInput.CharLimit = 100
	blocksInput.Width = 40
	blocksInput.Prompt = "> "

	// BlockedBy input
	blockedByInput := textinput.New()
	blockedByInput.Placeholder = "Task IDs (comma-separated, e.g. 1,2,3)"
	blockedByInput.CharLimit = 100
	blockedByInput.Width = 40
	blockedByInput.Prompt = "> "

	// Picker search input
	pickerSearch := textinput.New()
	pickerSearch.Placeholder = "Type to search tasks..."
	pickerSearch.CharLimit = 50
	pickerSearch.Width = 40
	pickerSearch.Prompt = "/ "

	// Statuses
	statuses := []string{"pending", "in_progress", "completed"}

	// Groups
	groups := append([]string{""}, groupStore.GetGroupNames()...)

	m := EditModel{
		taskStore:      taskStore,
		groupStore:     groupStore,
		isNew:          isNew,
		subjectInput:   subjectInput,
		descInput:      descInput,
		ownerInput:     ownerInput,
		blocksInput:    blocksInput,
		blockedByInput: blockedByInput,
		statuses:       statuses,
		groups:         groups,
		pickerSearch:   pickerSearch,
		pickerSelected: make(map[string]bool),
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
		m.blocksInput.SetValue(strings.Join(task.Blocks, ", "))
		m.blockedByInput.SetValue(strings.Join(task.BlockedBy, ", "))

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

	// Handle picker mode
	if m.pickerActive {
		return m.updatePicker(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s", "ctrl+enter":
			return m, m.save()
		case "esc":
			return m, func() tea.Msg {
				return CancelEditMsg{}
			}
		case "/":
			// Open picker for blocks/blockedBy fields
			if m.focusIdx == 5 || m.focusIdx == 6 {
				m.openPicker(m.focusIdx)
				return m, textinput.Blink
			}
		case "tab", "shift+tab":
			// Navigate fields (7 fields: 0-6)
			if msg.String() == "tab" {
				m.focusIdx = (m.focusIdx + 1) % 7
			} else {
				m.focusIdx = (m.focusIdx + 6) % 7
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
	case 5:
		m.blocksInput, cmd = m.blocksInput.Update(msg)
	case 6:
		m.blockedByInput, cmd = m.blockedByInput.Update(msg)
	}

	return m, cmd
}

func (m *EditModel) updateFocus() {
	m.subjectInput.Blur()
	m.descInput.Blur()
	m.ownerInput.Blur()
	m.blocksInput.Blur()
	m.blockedByInput.Blur()

	switch m.focusIdx {
	case 0:
		m.subjectInput.Focus()
	case 1:
		m.descInput.Focus()
	case 4:
		m.ownerInput.Focus()
	case 5:
		m.blocksInput.Focus()
	case 6:
		m.blockedByInput.Focus()
	}
}

func (m *EditModel) openPicker(field int) {
	m.pickerActive = true
	m.pickerForField = field
	m.pickerSearch.SetValue("")
	m.pickerSearch.Focus()
	m.pickerCursor = 0

	// Initialize selected from current field value
	m.pickerSelected = make(map[string]bool)
	var currentIDs []string
	if field == 5 {
		currentIDs = parseTaskIDs(m.blocksInput.Value())
	} else {
		currentIDs = parseTaskIDs(m.blockedByInput.Value())
	}
	for _, id := range currentIDs {
		m.pickerSelected[id] = true
	}

	m.filterPickerTasks()
}

func (m *EditModel) filterPickerTasks() {
	query := strings.ToLower(m.pickerSearch.Value())
	m.pickerTasks = nil

	for _, task := range m.taskStore.Tasks {
		// Skip self
		if !m.isNew && task.ID == m.task.ID {
			continue
		}

		// Filter by query
		if query != "" {
			if !strings.Contains(strings.ToLower(task.Subject), query) &&
				!strings.Contains(task.ID, query) {
				continue
			}
		}

		m.pickerTasks = append(m.pickerTasks, task)
	}

	// Reset cursor if out of bounds
	if m.pickerCursor >= len(m.pickerTasks) {
		m.pickerCursor = len(m.pickerTasks) - 1
	}
	if m.pickerCursor < 0 {
		m.pickerCursor = 0
	}
}

func (m EditModel) updatePicker(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.pickerActive = false
			m.updateFocus()
			return m, nil
		case "enter":
			// Toggle selection
			if len(m.pickerTasks) > 0 && m.pickerCursor < len(m.pickerTasks) {
				id := m.pickerTasks[m.pickerCursor].ID
				m.pickerSelected[id] = !m.pickerSelected[id]
			}
			return m, nil
		case "tab":
			// Confirm and close picker
			m.applyPickerSelection()
			m.pickerActive = false
			m.updateFocus()
			return m, nil
		case "up":
			if m.pickerCursor > 0 {
				m.pickerCursor--
			}
			return m, nil
		case "down":
			if m.pickerCursor < len(m.pickerTasks)-1 {
				m.pickerCursor++
			}
			return m, nil
		}
	}

	// Update search input
	m.pickerSearch, cmd = m.pickerSearch.Update(msg)
	m.filterPickerTasks()

	return m, cmd
}

func (m *EditModel) applyPickerSelection() {
	var ids []string
	for id, selected := range m.pickerSelected {
		if selected {
			ids = append(ids, id)
		}
	}
	value := strings.Join(ids, ", ")

	if m.pickerForField == 5 {
		m.blocksInput.SetValue(value)
	} else {
		m.blockedByInput.SetValue(value)
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

	// Parse blocks
	m.task.Blocks = parseTaskIDs(m.blocksInput.Value())
	m.task.BlockedBy = parseTaskIDs(m.blockedByInput.Value())

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

// SetSize updates the model dimensions and input widths
func (m *EditModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Skip updating input widths if model is not yet initialized
	if m.taskStore == nil {
		return
	}

	// Update input widths based on terminal width
	inputWidth := width - 6 // margin for borders and prompt
	if inputWidth < 40 {
		inputWidth = 40
	}
	m.subjectInput.Width = inputWidth
	m.descInput.SetWidth(inputWidth)
	m.ownerInput.Width = inputWidth
	m.blocksInput.Width = inputWidth
	m.blockedByInput.Width = inputWidth
	m.pickerSearch.Width = inputWidth
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

	// Picker overlay
	if m.pickerActive {
		return m.renderPicker()
	}

	// Subject field
	if m.focusIdx == 0 {
		b.WriteString(ui.SelectedStyle.Render("Subject:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Subject:"))
	}
	b.WriteString("\n")
	b.WriteString(m.subjectInput.View())
	b.WriteString("\n\n")

	// Description field
	if m.focusIdx == 1 {
		b.WriteString(ui.SelectedStyle.Render("Description:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Description:"))
	}
	b.WriteString("\n")
	b.WriteString(m.descInput.View())
	b.WriteString("\n\n")

	// Status selector
	if m.focusIdx == 2 {
		b.WriteString(ui.SelectedStyle.Render("Status:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Status:"))
	}
	b.WriteString(" ")

	statusText := m.statuses[m.statusIdx]
	statusIcon := ui.StatusIcon(statusText)
	statusStyle := ui.GetStatusStyle(statusText)
	if m.focusIdx == 2 {
		b.WriteString(statusStyle.Render(fmt.Sprintf("[%s %s] ↑↓", statusIcon, statusText)))
	} else {
		b.WriteString(statusStyle.Render(fmt.Sprintf(" %s %s", statusIcon, statusText)))
	}
	b.WriteString("\n\n")

	// Group selector
	if m.focusIdx == 3 {
		b.WriteString(ui.SelectedStyle.Render("Group:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Group:"))
	}
	b.WriteString(" ")

	groupText := "(none)"
	if m.groupIdx > 0 && m.groupIdx < len(m.groups) {
		groupText = m.groups[m.groupIdx]
	}

	if m.focusIdx == 3 {
		b.WriteString(fmt.Sprintf("[%s] ↑↓", groupText))
	} else {
		b.WriteString(fmt.Sprintf(" %s", groupText))
	}
	b.WriteString("\n\n")

	// Owner field
	if m.focusIdx == 4 {
		b.WriteString(ui.SelectedStyle.Render("Owner:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Owner:"))
	}
	b.WriteString("\n")
	b.WriteString(m.ownerInput.View())
	b.WriteString("\n\n")

	// Blocks field
	if m.focusIdx == 5 {
		b.WriteString(ui.SelectedStyle.Render("Blocks:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Blocks:"))
	}
	b.WriteString(ui.MutedStyle.Render(" (tasks that wait for this)"))
	b.WriteString("\n")
	b.WriteString(m.blocksInput.View())
	b.WriteString("\n\n")

	// BlockedBy field
	if m.focusIdx == 6 {
		b.WriteString(ui.SelectedStyle.Render("Blocked By:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Blocked By:"))
	}
	b.WriteString(ui.MutedStyle.Render(" (tasks this waits for)"))
	b.WriteString("\n")
	b.WriteString(m.blockedByInput.View())
	b.WriteString("\n")

	// Footer
	b.WriteString("\n")
	var keys [][]string
	if m.focusIdx == 5 || m.focusIdx == 6 {
		keys = [][]string{
			{"Tab", "Next Field"},
			{"/", "Search Tasks"},
			{"Ctrl+S", "Save"},
			{"Esc", "Cancel"},
		}
	} else {
		keys = [][]string{
			{"Tab", "Next Field"},
			{"Ctrl+S", "Save"},
			{"Esc", "Cancel"},
		}
	}
	b.WriteString(ui.Footer(keys, m.width))

	return b.String()
}

func (m EditModel) renderPicker() string {
	var b strings.Builder

	// Header
	fieldName := "Blocks"
	if m.pickerForField == 6 {
		fieldName = "Blocked By"
	}
	b.WriteString(ui.Header(fmt.Sprintf("Select Tasks for %s", fieldName), m.width))
	b.WriteString("\n\n")

	// Search
	b.WriteString(ui.InputLabelStyle.Render("Search:"))
	b.WriteString("\n")
	b.WriteString(m.pickerSearch.View())
	b.WriteString("\n\n")

	// Task list
	b.WriteString(ui.HorizontalLine(m.width))
	b.WriteString("\n")

	if len(m.pickerTasks) == 0 {
		b.WriteString(ui.MutedStyle.Render("No tasks found."))
		b.WriteString("\n")
	} else {
		maxVisible := 10
		startIdx := 0
		if m.pickerCursor >= maxVisible {
			startIdx = m.pickerCursor - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.pickerTasks) {
			endIdx = len(m.pickerTasks)
		}

		for i := startIdx; i < endIdx; i++ {
			task := m.pickerTasks[i]
			prefix := "  "
			if i == m.pickerCursor {
				prefix = "> "
			}

			checkbox := "[ ]"
			if m.pickerSelected[task.ID] {
				checkbox = "[✓]"
			}

			statusIcon := data.StatusIcon(task.Status)
			line := fmt.Sprintf("%s%s #%s %s %s", prefix, checkbox, task.ID, statusIcon, task.Subject)

			if i == m.pickerCursor {
				b.WriteString(ui.SelectedStyle.Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
	}

	// Footer
	b.WriteString("\n")
	keys := [][]string{
		{"↑↓", "Navigate"},
		{"Enter", "Toggle"},
		{"Tab", "Confirm"},
		{"Esc", "Cancel"},
	}
	b.WriteString(ui.Footer(keys, m.width))

	return b.String()
}

// parseTaskIDs parses comma-separated task IDs
func parseTaskIDs(input string) []string {
	if strings.TrimSpace(input) == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	var ids []string
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}
