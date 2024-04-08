package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type Action int

const (
	NULL Action = iota
	NEXT
)

type TeaTicker struct {
	Flag Action
}

func (t *TeaTicker) Update() tea.Cmd {
	return tea.Tick(600*time.Millisecond, func(time time.Time) tea.Msg {
		return t
	})
}
