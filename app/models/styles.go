package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// List Model Styles
var containerStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
var containerSelectedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("3"))
var containerNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Render
var containerIdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Render
var containerImageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render

// Pagenator Styles
var activeDot = lipgloss.NewStyle().Foreground(
	lipgloss.AdaptiveColor{Light: "235", Dark: "252"},
).Render("•")

var inactiveDot = lipgloss.NewStyle().Foreground(
	lipgloss.AdaptiveColor{Light: "250", Dark: "238"},
).Render("•")
