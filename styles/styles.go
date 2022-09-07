package styles

import (
	lp "github.com/charmbracelet/lipgloss"
)

var DoubleBorder = lp.NewStyle().Border(lp.DoubleBorder())
var DoubleBorderPadded = lp.NewStyle().Border(lp.DoubleBorder()).Padding(6)
var Blinking = lp.NewStyle().Blink(true)
var LeftPadding = lp.NewStyle().PaddingLeft(5)
