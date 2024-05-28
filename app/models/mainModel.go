package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/client"
)

// Modal state to determine which component is loaded
type state int

const (
	DefaultState state = iota
	ListState
	ViewLogState
)

// when the window size changes
type windowSizeUpdateEvent struct {
	Width  int
	Height int
}

// Update the view model information every second
type UpdateViewModel int

// determines if docker client is available
type dockerClientInitalized int

const (
	clientInit dockerClientInitalized = iota
	clientUninit
)

func initalizeDockerClient(m *MainModel) tea.Msg {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return dockerClientInitalized(clientUninit)
	}
	m.dockerClient = cli
	return dockerClientInitalized(clientInit)
}

type ViewStates interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	extendedHelpView() string
	helpView() string
}
type MainModel struct {
	viewState        state
	viewModel        ViewStates
	dockerClient     *client.Client
	width            int
	height           int
	extendedHelpFlag bool
}

// Update Containers List every second
func (m *MainModel) intervalUpdates() tea.Cmd {
	return tea.Tick(time.Second,
		func(_ time.Time) tea.Msg {
			return UpdateViewModel(0)
		},
	)
}
func (m *MainModel) Init() tea.Cmd {
	return func() tea.Msg { return initalizeDockerClient(m) }
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case dockerClientInitalized:
		if msg == clientInit {
			m.viewState = ListState
			m.viewModel = &ListModel{Client: m.dockerClient}
			cmd := m.viewModel.Init()
			return m, cmd
		}
		return m, tea.Quit
	case UpdateViewModel:
		m.viewModel.Update(msg)
		cmd = m.intervalUpdates()
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.viewState > ListState {
				m.viewState--
			} else {
				m.dockerClient.Close()
				return m, tea.Quit
			}
		case "?":
			m.extendedHelpFlag = !m.extendedHelpFlag
			if m.extendedHelpFlag {
				m.viewModel.Update(windowSizeUpdateEvent{Width: m.width, Height: m.height - 10})
				return m, nil
			}
			m.viewModel.Update(windowSizeUpdateEvent{Width: m.width, Height: m.height - 2})
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewModel.Update(windowSizeUpdateEvent{Width: msg.Width, Height: msg.Height})
		return m, nil
	}

	m.viewModel.Update(msg)
	return m, cmd
}

func (m *MainModel) View() string {
	if m.viewState == 0 {
		return "Loading ..."
	}
	if m.extendedHelpFlag {
		return m.viewModel.View() + m.viewModel.extendedHelpView()
	}
	return m.viewModel.View() + m.viewModel.helpView()
}
