package model

import (
	"github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
)

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
	return a.projects.Init()
}

// Update handles messages
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		// Reload tasks to reflect any changes
		a.taskStore, _ = data.LoadTasks(a.projectName)
		a.groupStore, _ = data.LoadGroups(a.projectName)
		a.tasks = NewTasksModel(a.projectName, a.taskStore, a.groupStore)
		a.tasks.width = a.width
		a.tasks.height = a.height
		a.screen = ScreenTasks
		return a, a.tasks.Init()

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
	if a.err != nil {
		return "Error: " + a.err.Error()
	}

	switch a.screen {
	case ScreenProjects:
		return a.projects.View()
	case ScreenTasks:
		return a.tasks.View()
	case ScreenDetail:
		return a.detail.View()
	case ScreenEdit:
		return a.edit.View()
	case ScreenGroups:
		return a.groups.View()
	case ScreenGroupEdit:
		return a.groupEdit.View()
	default:
		return "Unknown screen"
	}
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
