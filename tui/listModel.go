package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ListModel struct {
	Client           *client.Client
	pg               paginator.Model
	vp               viewport.Model
	items            []types.Container
	cursorPos        int
	width            int
	height           int
	extendedHelpFlag bool
}

type dockerClientListInitalized []types.Container

func dockerContianersList(m *ListModel) ([]types.Container, error) {
	containers, err := m.Client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (m *ListModel) Init() tea.Cmd {

	m.vp = viewport.Model{}
	pg := paginator.New()
	pg.Type = paginator.Dots
	pg.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	pg.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	m.pg = pg
	m.items, _ = dockerContianersList(m)
	m.pg.SetTotalPages(len(m.items))
	return m.intervalUpdates()
}
func (m *ListModel) intervalUpdates() tea.Cmd {
	return tea.Tick(time.Second,
		func(_ time.Time) tea.Msg {
			m.items, _ = dockerContianersList(m)
			m.pg.SetTotalPages(len(m.items))
			return dockerClientListInitalized(m.items)
		},
	)
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
		case "?":
			m.extendedHelpFlag = !m.extendedHelpFlag
			if m.extendedHelpFlag {
				m.vp.Height = m.height - 10
				return m, nil
			}
			m.vp.Height = m.height - 2
			return m, nil
		case "r":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerRestart(context.Background(), m.items[start+m.cursorPos].ID, container.StopOptions{})
		case "s":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerStart(context.Background(), m.items[start+m.cursorPos].ID, container.StartOptions{})
		case "p":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerStop(context.Background(), m.items[start+m.cursorPos].ID, container.StopOptions{})
		case "x":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerKill(context.Background(), m.items[start+m.cursorPos].ID, "SIGKILL")
		case "d":
			start, _ := m.pg.GetSliceBounds(len(m.items))
			m.Client.ContainerRemove(context.Background(), m.items[start+m.cursorPos].ID, container.RemoveOptions{RemoveVolumes: true, Force: true})

		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.vp.Height = msg.Height - 2
		m.vp.Width = msg.Width
		m.pg.PerPage = (msg.Height - 2) / 4
	case dockerClientListInitalized:
		m.items = msg
		cmd = m.intervalUpdates()
		return m, cmd
	}
	itemsOnPage := m.pg.ItemsOnPage(len(m.items))
	if m.cursorPos > itemsOnPage-1 {
		m.cursorPos = max(0, itemsOnPage-1)
	}
	m.pg, cmd = m.pg.Update(msg)
	return m, cmd
}

var containerStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
var containerSelectedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("3"))
var containerNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Render
var containerIdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Render
var containerImageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render

func (m *ListModel) View() string {
	var outputStr string
	start, end := m.pg.GetSliceBounds(len(m.items))
	for index, v := range m.items[start:end] {
		if m.cursorPos == index {
			outputStr += "\n" + containerSelectedStyle.Width(m.width-2).Render(containerString(v))
			continue
		}
		outputStr += "\n" + containerStyle.Width(m.width-2).Render(containerString(v))
	}
	m.vp.SetContent(outputStr)
	if m.extendedHelpFlag {
		return m.vp.View() + "\n" + m.pg.View() + extendedHelpView()
	}
	return m.vp.View() + "\n" + m.pg.View() + helpView()
}

func containerString(c types.Container) string {
	return fmt.Sprintf("%s \n %.24s\t %s\t %s",
		containerNameStyle(c.Names[0]),
		containerIdStyle(c.ID),
		containerImageStyle(c.Image),
		c.Status,
	)
}
func extendedHelpView() string {
	return "\n↑/↓ to navigate,\n r restart container, s start contianer, p pause, d destroy container x kill contianer, q to quit"
}
func helpView() string {
	return "\n↑/↓ to navigate, r restart, s start, p pause, q quit, d destroy, ? help"
}
