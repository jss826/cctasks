package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
	"github.com/jss826/cctasks/internal/ui"
)

// GroupsModel handles the group management screen
type GroupsModel struct {
	groupStore *data.GroupStore
	width      int
	height     int

	cursor        int
	confirmDelete bool
}

// NewGroupsModel creates a new GroupsModel
func NewGroupsModel(groupStore *data.GroupStore) GroupsModel {
	return GroupsModel{
		groupStore: groupStore,
	}
}

// Init initializes the model
func (m GroupsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m GroupsModel) Update(msg tea.Msg) (GroupsModel, tea.Cmd) {
	// Delete confirmation mode
	if m.confirmDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if len(m.groupStore.Groups) > 0 {
					groupName := m.groupStore.Groups[m.cursor].Name
					m.groupStore.DeleteGroup(groupName)
					m.groupStore.Save()
					if m.cursor >= len(m.groupStore.Groups) {
						m.cursor = len(m.groupStore.Groups) - 1
					}
					if m.cursor < 0 {
						m.cursor = 0
					}
				}
				m.confirmDelete = false
			case "n", "N", "esc":
				m.confirmDelete = false
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
			if m.cursor < len(m.groupStore.Groups)-1 {
				m.cursor++
			}
		case "enter", "e":
			if len(m.groupStore.Groups) > 0 {
				group := &m.groupStore.Groups[m.cursor]
				return m, func() tea.Msg {
					return EditGroupMsg{Group: group, IsNew: false}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return EditGroupMsg{Group: nil, IsNew: true}
			}
		case "d":
			if len(m.groupStore.Groups) > 0 {
				m.confirmDelete = true
			}
		case "K", "shift+up":
			// Move group up
			if len(m.groupStore.Groups) > 0 {
				name := m.groupStore.Groups[m.cursor].Name
				if m.groupStore.MoveGroupUp(name) {
					m.groupStore.Save()
					if m.cursor > 0 {
						m.cursor--
					}
				}
			}
		case "J", "shift+down":
			// Move group down
			if len(m.groupStore.Groups) > 0 {
				name := m.groupStore.Groups[m.cursor].Name
				if m.groupStore.MoveGroupDown(name) {
					m.groupStore.Save()
					if m.cursor < len(m.groupStore.Groups)-1 {
						m.cursor++
					}
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return BackFromGroupsMsg{}
			}
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the group management screen
func (m GroupsModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.Header("Groups", m.width))
	b.WriteString("\n\n")

	// Delete confirmation
	if m.confirmDelete && len(m.groupStore.Groups) > 0 {
		groupName := m.groupStore.Groups[m.cursor].Name
		dialog := ui.Confirm(
			"Delete Group",
			fmt.Sprintf("Are you sure you want to delete group \"%s\"?", groupName),
			"y", "n",
		)
		b.WriteString(dialog)
		b.WriteString("\n\n")
	}

	// Group list
	if len(m.groupStore.Groups) == 0 {
		b.WriteString(ui.MutedStyle.Render("No groups defined."))
		b.WriteString("\n")
		b.WriteString(ui.MutedStyle.Render("Press 'n' to create a new group."))
		b.WriteString("\n")
	}

	for i, group := range m.groupStore.Groups {
		prefix := "  "
		style := ui.NormalStyle
		if i == m.cursor {
			prefix = "> "
			style = ui.SelectedStyle
		}

		swatch := ui.ColorSwatchStyle(group.Color).Render("██")
		line := fmt.Sprintf("%s%s %s", prefix, swatch, group.Name)
		b.WriteString(style.Render(line))

		// Show move indicators
		moveHint := ""
		if i == m.cursor {
			if i > 0 {
				moveHint += " [K↑]"
			}
			if i < len(m.groupStore.Groups)-1 {
				moveHint += " [J↓]"
			}
		}
		b.WriteString(ui.MutedStyle.Render(moveHint))
		b.WriteString("\n")
	}

	// Add group option
	b.WriteString("\n")
	b.WriteString(ui.MutedStyle.Render("  [+ Add Group]"))
	b.WriteString("\n")

	// Footer
	b.WriteString("\n")
	keys := [][]string{
		{"↑↓", "Select"},
		{"Enter", "Edit"},
		{"n", "New"},
		{"d", "Delete"},
		{"K/J", "Reorder"},
		{"Esc", "Back"},
	}
	b.WriteString(ui.Footer(keys, m.width))

	return ui.AppStyle.Render(b.String())
}

// GroupEditModel handles the group edit dialog
type GroupEditModel struct {
	group      *data.TaskGroup
	groupStore *data.GroupStore
	isNew      bool
	width      int
	height     int

	nameInput   textinput.Model
	colorIdx    int
	focusIdx    int // 0=name, 1=color
}

// NewGroupEditModel creates a new GroupEditModel
func NewGroupEditModel(group *data.TaskGroup, groupStore *data.GroupStore, isNew bool) GroupEditModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Group name"
	nameInput.CharLimit = 50
	nameInput.Width = 40
	nameInput.Prompt = "> "
	nameInput.Focus()

	m := GroupEditModel{
		groupStore: groupStore,
		isNew:      isNew,
		nameInput:  nameInput,
	}

	if isNew {
		m.group = &data.TaskGroup{
			Color: data.DefaultColors[0],
		}
	} else {
		// Copy existing group
		groupCopy := *group
		m.group = &groupCopy
		m.nameInput.SetValue(group.Name)

		// Find color index
		for i, c := range data.DefaultColors {
			if c == group.Color {
				m.colorIdx = i
				break
			}
		}
	}

	return m
}

// Init initializes the model
func (m GroupEditModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m GroupEditModel) Update(msg tea.Msg) (GroupEditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "ctrl+s":
			return m, m.save()
		case "esc":
			return m, func() tea.Msg {
				return CancelGroupEditMsg{}
			}
		case "tab":
			m.focusIdx = (m.focusIdx + 1) % 2
			if m.focusIdx == 0 {
				m.nameInput.Focus()
			} else {
				m.nameInput.Blur()
			}
			return m, nil
		case "left":
			if m.focusIdx == 1 && m.colorIdx > 0 {
				m.colorIdx--
			}
			return m, nil
		case "right":
			if m.focusIdx == 1 && m.colorIdx < len(data.DefaultColors)-1 {
				m.colorIdx++
			}
			return m, nil
		}
	}

	if m.focusIdx == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	}

	return m, cmd
}

func (m *GroupEditModel) save() tea.Cmd {
	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		return nil
	}

	color := data.DefaultColors[m.colorIdx]

	if m.isNew {
		m.groupStore.AddGroup(data.TaskGroup{
			Name:  name,
			Color: color,
		})
	} else {
		oldName := m.group.Name
		m.groupStore.UpdateGroup(oldName, data.TaskGroup{
			Name:  name,
			Order: m.group.Order,
			Color: color,
		})
	}
	m.groupStore.Save()

	return func() tea.Msg {
		return GroupSavedMsg{Store: m.groupStore}
	}
}

// View renders the group edit dialog
func (m GroupEditModel) View() string {
	var b strings.Builder

	// Header
	title := "New Group"
	if !m.isNew {
		title = "Edit Group"
	}
	b.WriteString(ui.Header(title, m.width))
	b.WriteString("\n\n")

	// Name field
	if m.focusIdx == 0 {
		b.WriteString(ui.SelectedStyle.Render("Name:"))
	} else {
		b.WriteString(ui.InputLabelStyle.Render("Name:"))
	}
	b.WriteString("\n")
	b.WriteString(m.nameInput.View())
	b.WriteString("\n\n")

	// Color field
	colorLabel := ui.InputLabelStyle.Render("Color:")
	if m.focusIdx == 1 {
		colorLabel = ui.SelectedStyle.Render("Color:")
	}
	b.WriteString(colorLabel)
	b.WriteString(" ")

	currentColor := data.DefaultColors[m.colorIdx]
	b.WriteString(ui.ColorSwatchStyle(currentColor).Render("████"))
	b.WriteString(" " + currentColor)
	b.WriteString("\n\n")

	// Color palette
	b.WriteString(ui.MutedStyle.Render("Preset Colors:"))
	b.WriteString("\n")
	for i, color := range data.DefaultColors {
		swatch := ui.ColorSwatchStyle(color).Render("██")
		if i == m.colorIdx && m.focusIdx == 1 {
			b.WriteString("[" + swatch + "]")
		} else {
			b.WriteString(" " + swatch + " ")
		}
	}
	b.WriteString("\n")

	// Footer
	b.WriteString("\n")
	keys := [][]string{
		{"Tab", "Next"},
		{"←→", "Color"},
		{"Enter", "Save"},
		{"Esc", "Cancel"},
	}
	b.WriteString(ui.Footer(keys, m.width))

	return ui.AppStyle.Render(b.String())
}
