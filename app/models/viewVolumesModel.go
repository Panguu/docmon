package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type ViewVolumesModel struct {
	Client    *client.Client
	pg        paginator.Model
	vp        viewport.Model
	items     []*volume.Volume
	cursorPos int
	width     int
	height    int
}

// Get Contianer List
func volumeContianersList(m *ViewVolumesModel) ([]*volume.Volume, error) {
	volume, err := m.Client.VolumeList(
		context.Background(),
		volume.ListOptions{},
	)
	if err != nil {
		return nil, err
	}
	return volume.Volumes, nil
}

func (m *ViewVolumesModel) Init() tea.Cmd {

	m.vp = viewport.Model{}
	pg := paginator.New()
	pg.Type = paginator.Dots
	pg.ActiveDot = activeDot
	pg.InactiveDot = inactiveDot
	m.pg = pg
	m.items, _ = volumeContianersList(m)
	m.pg.SetTotalPages(len(m.items))
	return func() tea.Msg { return UpdateViewModel(0) }
}

func (m *ViewVolumesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		}

	case windowSizeUpdateEvent:
		m.width = msg.Width
		m.height = msg.Height
		m.vp.Height = msg.Height - 2
		m.vp.Width = msg.Width
		m.pg.PerPage = (msg.Height - 2) / 4

	case UpdateViewModel:
		m.items, _ = volumeContianersList(m)
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

func (m *ViewVolumesModel) View() string {
	var outputStr string
	start, end := m.pg.GetSliceBounds(len(m.items))
	for index, v := range m.items[start:end] {
		if m.cursorPos == index {
			outputStr += "\n" + containerSelectedStyle.Width(m.width-2).Render(volumeString(v))
			continue
		}
		outputStr += "\n" + containerStyle.
			Width(m.width-2).
			Render(volumeString(v))
	}

	m.vp.SetContent(outputStr)

	return m.vp.View() + "\n" + m.pg.View()
}

func volumeString(c *volume.Volume) string {
	return fmt.Sprintf("%s \n %.24s\t %s\t %s",
		containerNameStyle(c.Name),
		containerIdStyle(c.CreatedAt),
		containerImageStyle(c.Scope),
		c.Status,
	)
}
func (*ViewVolumesModel) extendedHelpView() string {
	return "\n↑/↓ to navigate,\n r restart container, s start contianer, p pause, d destroy container x kill contianer, q to quit"
}
func (*ViewVolumesModel) helpView() string {
	return "\n↑/↓ to navigate, r restart, s start, p pause, q quit, d destroy, ? help"
}
