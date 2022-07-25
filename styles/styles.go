package styles

import (
	lp "github.com/charmbracelet/lipgloss"
)

var DoubleBorder = lp.NewStyle().Border(lp.DoubleBorder())
var Blinking = lp.NewStyle().Blink(true)
var LeftPadding = lp.NewStyle().PaddingLeft(5)
