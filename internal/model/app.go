package model

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/jss826/cctasks/internal/data"
)

// checkSizeMsg is sent periodically to check for terminal resize (Windows workaround)
type checkSizeMsg struct{}

// checkSizeCmd returns a command that periodically checks terminal size
func checkSizeCmd() tea.Cmd {
	return tea.Tick(time.Second/10, func(t time.Time) tea.Msg {
		return checkSizeMsg{}
	})
}

// AppVersion is set from main.go
var AppVersion = "dev"

// Screen represents the current screen
type Screen int

const (
	ScreenProjects Screen = iota
	ScreenTasks
	ScreenDetail
	ScreenEdit
	ScreenGroups
	ScreenGroupEdit
)

// App is the main application model
type App struct {
	screen      Screen
	prevScreen  Screen
	width       int
	height      int
	projectName string

	// Sub-models
	projects  ProjectsModel
	tasks     TasksModel
	detail    DetailModel
	edit      EditModel
	groups    GroupsModel
	groupEdit GroupEditModel

	// Shared data
	taskStore  *data.TaskStore
	groupStore *data.GroupStore

	// State
	err error
}

// NewApp creates a new App model
func NewApp() App {
	return App{
		screen:   ScreenProjects,
		projects: NewProjectsModel(),
	}
}

// Init initializes the application
func (a App) Init() tea.Cmd {
	return tea.Batch(a.projects.Init(), checkSizeCmd())
}

// Update handles messages
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case checkSizeMsg:
		// Poll terminal size (Windows workaround for no SIGWINCH)
		// Try stdin first (works better with alt screen), fallback to stdout
		fd := int(os.Stdin.Fd())
		w, h, err := term.GetSize(fd)
		if err != nil {
			fd = int(os.Stdout.Fd())
			w, h, err = term.GetSize(fd)
		}
		if err != nil {
			return a, checkSizeCmd()
		}
		if w != a.width || h != a.height {
			a.width = w
			a.height = h
			// Propagate to sub-models
			a.projects.width = w
			a.projects.height = h
			a.tasks.width = w
			a.tasks.height = h
			a.detail.width = w
			a.detail.height = h
			a.edit.width = w
			a.edit.height = h
			a.groups.width = w
			a.groups.height = h
			a.groupEdit.width = w
			a.groupEdit.height = h
			// Clear screen and continue polling
			return a, tea.Batch(
				func() tea.Msg { return tea.ClearScreen() },
				checkSizeCmd(),
			)
		}
		return a, checkSizeCmd()

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to sub-models
		a.projects.width = msg.Width
		a.projects.height = msg.Height
		a.tasks.width = msg.Width
		a.tasks.height = msg.Height
		a.detail.width = msg.Width
		a.detail.height = msg.Height
		a.edit.width = msg.Width
		a.edit.height = msg.Height
		a.groups.width = msg.Width
		a.groups.height = msg.Height
		a.groupEdit.width = msg.Width
		a.groupEdit.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "ctrl+l":
			// Manual screen refresh
			return a, func() tea.Msg { return tea.ClearScreen() }
		}

		// Auto-reload on any key press if data has changed
		// Skip reload on edit screens (Groups, GroupEdit, Edit) to avoid cursor/state reset
		if a.projectName != "" && a.taskStore != nil && a.screen != ScreenGroups && a.screen != ScreenGroupEdit && a.screen != ScreenEdit {
			needsReload := a.taskStore.NeedsReload()
			if a.groupStore != nil && a.groupStore.NeedsReload() {
				needsReload = true
			}
			if needsReload {
				a.taskStore, _ = data.LoadTasks(a.projectName)
				a.groupStore, _ = data.LoadGroups(a.projectName)
				// Update current screen's data
				switch a.screen {
				case ScreenTasks:
					a.tasks = NewTasksModel(a.projectName, a.taskStore, a.groupStore)
					a.tasks.width = a.width
					a.tasks.height = a.height
				}
			}
		}

	case SelectProjectMsg:
		a.projectName = msg.Name
		var err error
		a.taskStore, err = data.LoadTasks(a.projectName)
		if err != nil {
			a.err = err
			return a, nil
		}
		a.groupStore, err = data.LoadGroups(a.projectName)
		if err != nil {
			a.err = err
			return a, nil
		}
		a.tasks = NewTasksModel(a.projectName, a.taskStore, a.groupStore)
		a.tasks.width = a.width
		a.tasks.height = a.height
		a.screen = ScreenTasks
		return a, a.tasks.Init()

	case BackToProjectsMsg:
		a.screen = ScreenProjects
		return a, a.projects.Init()

	case ViewTaskMsg:
		a.detail = NewDetailModel(msg.Task, a.taskStore, a.groupStore)
		a.detail.width = a.width
		a.detail.height = a.height
		a.prevScreen = ScreenTasks
		a.screen = ScreenDetail
		return a, nil

	case BackToTasksMsg:
		// Reload tasks to reflect any changes, preserving UI state
		a.taskStore, _ = data.LoadTasks(a.projectName)
		a.groupStore, _ = data.LoadGroups(a.projectName)
		a.tasks.ReloadData(a.taskStore, a.groupStore)
		a.screen = ScreenTasks
		return a, nil

	case EditTaskMsg:
		a.edit = NewEditModel(msg.Task, a.taskStore, a.groupStore, false)
		a.edit.width = a.width
		a.edit.height = a.height
		a.prevScreen = a.screen
		a.screen = ScreenEdit
		return a, a.edit.Init()

	case NewTaskMsg:
		a.edit = NewEditModel(nil, a.taskStore, a.groupStore, true)
		a.edit.width = a.width
		a.edit.height = a.height
		a.prevScreen = a.screen
		a.screen = ScreenEdit
		return a, a.edit.Init()

	case TaskSavedMsg:
		a.taskStore = msg.Store
		a.tasks = NewTasksModel(a.projectName, a.taskStore, a.groupStore)
		a.tasks.width = a.width
		a.tasks.height = a.height
		a.screen = ScreenTasks
		return a, a.tasks.Init()

	case CancelEditMsg:
		if a.prevScreen == ScreenDetail {
			a.screen = ScreenDetail
		} else {
			a.screen = ScreenTasks
		}
		return a, nil

	case ManageGroupsMsg:
		a.groups = NewGroupsModel(a.groupStore)
		a.groups.width = a.width
		a.groups.height = a.height
		a.prevScreen = a.screen
		a.screen = ScreenGroups
		return a, a.groups.Init()

	case BackFromGroupsMsg:
		// Reload groups
		a.groupStore, _ = data.LoadGroups(a.projectName)
		a.tasks = NewTasksModel(a.projectName, a.taskStore, a.groupStore)
		a.tasks.width = a.width
		a.tasks.height = a.height
		a.screen = ScreenTasks
		return a, a.tasks.Init()

	case EditGroupMsg:
		a.groupEdit = NewGroupEditModel(msg.Group, a.groupStore, msg.IsNew)
		a.groupEdit.width = a.width
		a.groupEdit.height = a.height
		a.screen = ScreenGroupEdit
		return a, a.groupEdit.Init()

	case GroupSavedMsg:
		a.groupStore = msg.Store
		a.groups = NewGroupsModel(a.groupStore)
		a.groups.width = a.width
		a.groups.height = a.height
		a.screen = ScreenGroups
		return a, a.groups.Init()

	case CancelGroupEditMsg:
		a.screen = ScreenGroups
		return a, nil

	case RefreshMsg:
		// Reload data
		if a.projectName != "" {
			a.taskStore, _ = data.LoadTasks(a.projectName)
			a.groupStore, _ = data.LoadGroups(a.projectName)
			if a.screen == ScreenTasks {
				a.tasks = NewTasksModel(a.projectName, a.taskStore, a.groupStore)
				a.tasks.width = a.width
				a.tasks.height = a.height
				return a, a.tasks.Init()
			}
		}
		return a, nil
	}

	// Delegate to current screen
	var cmd tea.Cmd
	switch a.screen {
	case ScreenProjects:
		a.projects, cmd = a.projects.Update(msg)
	case ScreenTasks:
		a.tasks, cmd = a.tasks.Update(msg)
	case ScreenDetail:
		a.detail, cmd = a.detail.Update(msg)
	case ScreenEdit:
		a.edit, cmd = a.edit.Update(msg)
	case ScreenGroups:
		a.groups, cmd = a.groups.Update(msg)
	case ScreenGroupEdit:
		a.groupEdit, cmd = a.groupEdit.Update(msg)
	}

	return a, cmd
}

// View renders the application
func (a App) View() string {
	var content string

	if a.err != nil {
		content = "Error: " + a.err.Error()
	} else {
		switch a.screen {
		case ScreenProjects:
			content = a.projects.View()
		case ScreenTasks:
			content = a.tasks.View()
		case ScreenDetail:
			content = a.detail.View()
		case ScreenEdit:
			content = a.edit.View()
		case ScreenGroups:
			content = a.groups.View()
		case ScreenGroupEdit:
			content = a.groupEdit.View()
		default:
			content = "Unknown screen"
		}
	}

	return content
}

// Messages for screen transitions

type SelectProjectMsg struct {
	Name string
}

type BackToProjectsMsg struct{}

type ViewTaskMsg struct {
	Task *data.Task
}

type BackToTasksMsg struct{}

type EditTaskMsg struct {
	Task *data.Task
}

type NewTaskMsg struct{}

type TaskSavedMsg struct {
	Store *data.TaskStore
}

type CancelEditMsg struct{}

type ManageGroupsMsg struct{}

type BackFromGroupsMsg struct{}

type EditGroupMsg struct {
	Group *data.TaskGroup
	IsNew bool
}

type GroupSavedMsg struct {
	Store *data.GroupStore
}

type CancelGroupEditMsg struct{}

type RefreshMsg struct{}
