package main

import (
	"docmon/tui"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/docker/docker/client"
)

// Modal state to determine which component is loaded
type state int

const (
	defaultState state = iota
	listState
	viewLogState
)

// determines if docker client is available
type dockerClientInitalized int

const (
	clientInit dockerClientInitalized = iota
	clientUninit
)

type viewModelInit int

const (
	viewInit viewModelInit = iota
	viewUninit
)

func initalizeDockerClient(m *mainModel) tea.Msg {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return dockerClientInitalized(clientUninit)
	}
	m.dockerClient = cli
	return dockerClientInitalized(clientInit)
}

type ViewStates interface {
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}
type mainModel struct {
	viewState    state
	viewModel    tea.Model
	dockerClient *client.Client
}

func (m *mainModel) Init() tea.Cmd {
	return func() tea.Msg { return initalizeDockerClient(m) }
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.viewState > listState {
				m.viewState--
			} else {
				m.dockerClient.Close()
				return m, tea.Quit
			}
		}
	case dockerClientInitalized:
		if msg == clientInit {
			m.viewState = listState
			m.viewModel = &tui.ListModel{Client: m.dockerClient}
			cmd := m.viewModel.Init()
			return m, cmd
		}
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.viewModel, cmd = m.viewModel.Update(msg)
	return m, cmd
}

func (m *mainModel) View() string {
	switch m.viewState {
	case listState:
		return m.viewModel.View()
	default:
		return "Loading..."
	}

}

func main() {
	p := tea.NewProgram(&mainModel{viewState: defaultState}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
