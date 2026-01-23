package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jss826/cctasks/internal/data"
	"github.com/jss826/cctasks/internal/ui"
)

// TasksModel handles the task list screen
type TasksModel struct {
	projectName string
	taskStore   *data.TaskStore
	groupStore  *data.GroupStore
	width       int
	height      int

	// Navigation
	cursor int
	items  []taskListItem // Flattened list of groups and tasks

	// Filtering
	statusFilter string   // "", "pending", "in_progress", "completed"
	groupFilter  string   // "", or group name
	searchInput  textinput.Model
	searchActive bool

	// Group collapsed state
	collapsedGroups map[string]bool

	// Quick status change mode
	statusChangeMode bool
}

// taskListItem represents an item in the flattened task list
type taskListItem struct {
	isGroup   bool
	groupName string
	task      *data.Task
}

// NewTasksModel creates a new TasksModel
func NewTasksModel(projectName string, taskStore *data.TaskStore, groupStore *data.GroupStore) TasksModel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 50
	ti.Width = 30

	m := TasksModel{
		projectName:     projectName,
		taskStore:       taskStore,
		groupStore:      groupStore,
		searchInput:     ti,
		collapsedGroups: make(map[string]bool),
	}
	m.rebuildItems()
	return m
}

// Init initializes the model
func (m TasksModel) Init() tea.Cmd {
	return nil
}

// rebuildItems rebuilds the flattened list based on current filters
func (m *TasksModel) rebuildItems() {
	m.items = nil

	// Get filtered tasks
	var tasks []data.Task
	for _, task := range m.taskStore.Tasks {
		// Status filter
		if m.statusFilter != "" && task.Status != m.statusFilter {
			continue
		}

		// Group filter
		taskGroup := data.GetTaskGroup(task)
		if m.groupFilter != "" && taskGroup != m.groupFilter {
			continue
		}

		// Search filter
		if m.searchInput.Value() != "" {
			query := strings.ToLower(m.searchInput.Value())
			if !strings.Contains(strings.ToLower(task.Subject), query) &&
				!strings.Contains(strings.ToLower(task.Description), query) {
				continue
			}
		}

		tasks = append(tasks, task)
	}

	// Group tasks by group name
	groupedTasks := make(map[string][]data.Task)
	for _, task := range tasks {
		group := data.GetTaskGroup(task)
		if group == "" {
			group = "Uncategorized"
		}
		groupedTasks[group] = append(groupedTasks[group], task)
	}

	// Get group order from groupStore
	groupOrder := m.groupStore.GetGroupNames()

	// Add groups in order
	processedGroups := make(map[string]bool)

	for _, groupName := range groupOrder {
		if tasks, ok := groupedTasks[groupName]; ok {
			m.addGroupToItems(groupName, tasks)
			processedGroups[groupName] = true
		}
	}

	// Add remaining groups (including Uncategorized)
	for groupName, tasks := range groupedTasks {
		if !processedGroups[groupName] {
			m.addGroupToItems(groupName, tasks)
		}
	}

	// Ensure cursor is valid
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *TasksModel) addGroupToItems(groupName string, tasks []data.Task) {
	// Add group header
	m.items = append(m.items, taskListItem{
		isGroup:   true,
		groupName: groupName,
	})

	// Add tasks if not collapsed
	if !m.collapsedGroups[groupName] {
		for i := range tasks {
			m.items = append(m.items, taskListItem{
				task: &tasks[i],
			})
		}
	}
}

// Update handles messages
func (m TasksModel) Update(msg tea.Msg) (TasksModel, tea.Cmd) {
	var cmd tea.Cmd

	// Handle search input
	if m.searchActive {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searchActive = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				m.searchActive = false
				m.searchInput.Blur()
				m.rebuildItems()
				return m, nil
			}
		}
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.rebuildItems()
		return m, cmd
	}

	// Handle status change mode
	if m.statusChangeMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "1", "p":
				m.setCurrentTaskStatus("pending")
				m.statusChangeMode = false
			case "2", "i":
				m.setCurrentTaskStatus("in_progress")
				m.statusChangeMode = false
			case "3", "c":
				m.setCurrentTaskStatus("completed")
				m.statusChangeMode = false
			case "esc":
				m.statusChangeMode = false
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.items) > 0 {
				item := m.items[m.cursor]
				if item.isGroup {
					// Toggle collapse
					m.collapsedGroups[item.groupName] = !m.collapsedGroups[item.groupName]
					m.rebuildItems()
				} else if item.task != nil {
					return m, func() tea.Msg {
						return ViewTaskMsg{Task: item.task}
					}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return NewTaskMsg{}
			}
		case "e":
			if len(m.items) > 0 {
				item := m.items[m.cursor]
				if item.task != nil {
					return m, func() tea.Msg {
						return EditTaskMsg{Task: item.task}
					}
				}
			}
		case "s":
			if len(m.items) > 0 && m.items[m.cursor].task != nil {
				m.statusChangeMode = true
			}
		case "f":
			m.cycleStatusFilter()
			m.rebuildItems()
		case "g":
			m.cycleGroupFilter()
			m.rebuildItems()
		case "G":
			return m, func() tea.Msg {
				return ManageGroupsMsg{}
			}
		case "/":
			m.searchActive = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case "p":
			return m, func() tea.Msg {
				return BackToProjectsMsg{}
			}
		case "r":
			return m, func() tea.Msg {
				return RefreshMsg{}
			}
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *TasksModel) cycleStatusFilter() {
	statuses := []string{"", "pending", "in_progress", "completed"}
	for i, s := range statuses {
		if s == m.statusFilter {
			m.statusFilter = statuses[(i+1)%len(statuses)]
			return
		}
	}
	m.statusFilter = ""
}

func (m *TasksModel) cycleGroupFilter() {
	groups := append([]string{""}, m.groupStore.GetGroupNames()...)
	groups = append(groups, "Uncategorized")

	for i, g := range groups {
		if g == m.groupFilter {
			m.groupFilter = groups[(i+1)%len(groups)]
			return
		}
	}
	m.groupFilter = ""
}

func (m *TasksModel) setCurrentTaskStatus(status string) {
	if len(m.items) == 0 {
		return
	}
	item := m.items[m.cursor]
	if item.task == nil {
		return
	}

	item.task.Status = status
	m.taskStore.UpdateTask(*item.task)
	m.taskStore.Save()
	m.rebuildItems()
}

// View renders the task list screen
func (m TasksModel) View() string {
	var b strings.Builder

	// Header
	title := fmt.Sprintf("cctasks: %s", m.projectName)
	b.WriteString(ui.Header(title, m.width))
	b.WriteString("\n")

	// Filter bar
	statusLabel := "All"
	if m.statusFilter != "" {
		statusLabel = m.statusFilter
	}
	groupLabel := "All Groups"
	if m.groupFilter != "" {
		groupLabel = m.groupFilter
	}

	filterBar := fmt.Sprintf("Filter: [%s] [%s]  Search: [%s]",
		statusLabel, groupLabel, m.searchInput.View())
	b.WriteString(ui.FilterBarStyle.Render(filterBar))
	b.WriteString("\n")

	b.WriteString(ui.HorizontalLine(m.width - 4))
	b.WriteString("\n")

	// Status change mode indicator
	if m.statusChangeMode {
		b.WriteString(ui.WarningStyle.Render("Change status: [1/p] pending  [2/i] in_progress  [3/c] completed  [Esc] cancel"))
		b.WriteString("\n\n")
	}

	// Search mode indicator
	if m.searchActive {
		b.WriteString(ui.WarningStyle.Render("Search: Type to filter, [Enter] confirm, [Esc] cancel"))
		b.WriteString("\n\n")
	}

	// Task list
	if len(m.items) == 0 {
		b.WriteString(ui.MutedStyle.Render("No tasks found."))
		b.WriteString("\n")
		b.WriteString(ui.MutedStyle.Render("Press 'n' to create a new task."))
		b.WriteString("\n")
	}

	// Calculate visible area
	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 10
	}

	startIdx := 0
	if m.cursor >= maxVisible {
		startIdx = m.cursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}

	for i := startIdx; i < endIdx; i++ {
		item := m.items[i]
		isSelected := i == m.cursor

		if item.isGroup {
			b.WriteString(m.renderGroupHeader(item.groupName, isSelected))
		} else if item.task != nil {
			b.WriteString(m.renderTaskItem(item.task, isSelected))
		}
		b.WriteString("\n")
	}

	// Footer - context-aware hints
	b.WriteString("\n")

	// Check if a task is selected (not a group)
	taskSelected := false
	if len(m.items) > 0 && m.cursor < len(m.items) {
		taskSelected = m.items[m.cursor].task != nil
	}

	hints := []ui.KeyHint{
		{Key: "↑↓", Desc: "Navigate", Enabled: len(m.items) > 0},
		{Key: "Enter", Desc: "Select", Enabled: len(m.items) > 0},
		{Key: "n", Desc: "New", Enabled: true},
		{Key: "e", Desc: "Edit", Enabled: taskSelected},
		{Key: "s", Desc: "Status", Enabled: taskSelected},
		{Key: "f", Desc: "Filter", Enabled: true},
		{Key: "G", Desc: "Groups", Enabled: true},
		{Key: "/", Desc: "Search", Enabled: true},
		{Key: "Ctrl+L", Desc: "Redraw", Enabled: true},
		{Key: "p", Desc: "Projects", Enabled: true},
		{Key: "q", Desc: "Quit", Enabled: true},
	}
	b.WriteString(ui.FooterWithHints(hints, m.width))

	return ui.AppStyle.Render(b.String())
}

func (m *TasksModel) renderGroupHeader(groupName string, selected bool) string {
	// Get task count for this group
	count := 0
	for _, task := range m.taskStore.Tasks {
		tg := data.GetTaskGroup(task)
		if tg == "" {
			tg = "Uncategorized"
		}
		if tg == groupName {
			count++
		}
	}

	// Get group color
	color := m.groupStore.GetGroupColor(groupName)
	if groupName == "Uncategorized" {
		color = "#6b7280"
	}

	// Collapse indicator
	collapseIcon := "▼"
	if m.collapsedGroups[groupName] {
		collapseIcon = "▶"
	}

	prefix := "  "
	style := ui.GroupHeaderStyle
	if selected {
		prefix = "> "
		style = ui.SelectedStyle
	}

	swatch := ui.ColorSwatchStyle(color).Render("●")
	header := fmt.Sprintf("%s%s %s %s (%d)", prefix, collapseIcon, swatch, groupName, count)
	result := style.Render(header)

	// Show hint when selected
	if selected {
		hint := " (Enter: toggle)"
		result += ui.MutedStyle.Render(hint)
	}

	return result
}

func (m *TasksModel) renderTaskItem(task *data.Task, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}

	statusIcon := data.StatusIcon(task.Status)
	statusStyle := ui.GetStatusStyle(task.Status)
	statusBadge := statusStyle.Render(fmt.Sprintf("[%s]", task.Status))

	// Calculate available width for subject
	statusWidth := lipgloss.Width(statusBadge)
	maxSubjectLen := m.width - 25 - statusWidth
	if maxSubjectLen < 20 {
		maxSubjectLen = 20
	}
	subject := ui.Truncate(task.Subject, maxSubjectLen)

	// Build left part (without styling yet)
	leftContent := fmt.Sprintf("%s%s #%s %s",
		prefix,
		statusStyle.Render(statusIcon),
		task.ID,
		subject,
	)

	// Calculate padding using lipgloss.Width for accurate measurement
	leftWidth := lipgloss.Width(leftContent)
	totalWidth := m.width - 8
	if totalWidth < 60 {
		totalWidth = 80
	}
	padding := totalWidth - leftWidth - statusWidth
	if padding < 1 {
		padding = 1
	}

	line := leftContent + strings.Repeat(" ", padding) + statusBadge

	var result string
	if selected {
		result = ui.TaskSelectedStyle.Render(line)
	} else {
		result = ui.TaskItemStyle.Render(line)
	}

	// Add blocked by indicator
	if len(task.BlockedBy) > 0 {
		blockedByStr := fmt.Sprintf("      └─ blocked by: %s", strings.Join(task.BlockedBy, ", "))
		result += "\n" + ui.BlockedByStyle.Render(blockedByStr)
	}

	return result
}
