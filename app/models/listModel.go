package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// List Model Struct
type ListModel struct {
	Client    *client.Client
	pg        paginator.Model
	vp        viewport.Model
	items     []types.Container
	cursorPos int
	width     int
	height    int
}

// Get Contianer List
func dockerContianersList(m *ListModel) ([]types.Container, error) {
	containers, err := m.Client.ContainerList(
		context.Background(),
		container.ListOptions{All: true},
	)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (m *ListModel) Init() tea.Cmd {

	m.vp = viewport.Model{}
	pg := paginator.New()
	pg.Type = paginator.Dots
	pg.ActiveDot = activeDot
	pg.InactiveDot = inactiveDot
	m.pg = pg
	m.items, _ = dockerContianersList(m)
	m.pg.SetTotalPages(len(m.items))
	return func() tea.Msg { return UpdateViewModel(0) }
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "up":
			if m.cursorPos > 0 {
				m.cursorPos--
			}
		case "k", "down":
			if m.cursorPos < len(m.items)-1 {
				m.cursorPos++
			}

		case "r":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerRestart(
				context.Background(),
				m.items[start+m.cursorPos].ID,
				container.StopOptions{},
			)

		case "s":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerStart(
				context.Background(),
				m.items[start+m.cursorPos].ID,
				container.StartOptions{},
			)

		case "p":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerStop(
				context.Background(),
				m.items[start+m.cursorPos].ID,
				container.StopOptions{},
			)

		case "x":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerKill(
				context.Background(),
				m.items[start+m.cursorPos].ID,
				"SIGKILL",
			)

		case "d":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerRemove(
				context.Background(),
				m.items[start+m.cursorPos].ID,
				container.RemoveOptions{RemoveVolumes: true, Force: true},
			)

		}

	case windowSizeUpdateEvent:
		m.width = msg.Width
		m.height = msg.Height
		m.vp.Height = msg.Height - 2
		m.vp.Width = msg.Width
		m.pg.PerPage = (msg.Height - 2) / 4

	case UpdateViewModel:
		m.items, _ = dockerContianersList(m)
		m.pg.SetTotalPages(len(m.items))
		return m, cmd
	}

	itemsOnPage := m.pg.ItemsOnPage(len(m.items))
	if m.cursorPos > itemsOnPage-1 {
		m.cursorPos = max(0, itemsOnPage-1)
	}
	m.pg, cmd = m.pg.Update(msg)
	return m, cmd
}

func (m *ListModel) View() string {
	var outputStr string
	start, end := m.pg.GetSliceBounds(len(m.items))
	for index, v := range m.items[start:end] {
		if m.cursorPos == index {
			outputStr += "\n" + containerSelectedStyle.Width(m.width-2).Render(containerString(v))
			continue
		}
		outputStr += "\n" + containerStyle.
			Width(m.width-2).
			Render(containerString(v))
	}

	m.vp.SetContent(outputStr)

	return m.vp.View() + "\n" + m.pg.View()
}

func containerString(c types.Container) string {
	return fmt.Sprintf("%s \n %.24s\t %s\t %s",
		containerNameStyle(c.Names[0]),
		containerIdStyle(c.ID),
		containerImageStyle(c.Image),
		c.Status,
	)
}

func (*ListModel) extendedHelpView() string {
	return "\n↑/↓ to navigate,\n r restart container, s start contianer, p pause, d destroy container x kill contianer, q to quit"
}
func (*ListModel) helpView() string {
	return "\n↑/↓ to navigate, r restart, s start, p pause, q quit, d destroy, ? help"
}
