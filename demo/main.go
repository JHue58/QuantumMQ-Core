package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

type hello struct {
	ticker *TeaTicker

	name   string
	cursor int
	shane  string
}

func (h *hello) Init() tea.Cmd {
	h.ticker = &TeaTicker{Flag: NULL}
	h.shane = "_"
	return h.ticker.Update()
}

func (h *hello) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *TeaTicker:
		if msg.Flag == NULL {
			if h.shane == " " {
				h.shane = "_"
			} else {
				h.shane = " "
			}
			return h, msg.Update()
		} else {
			if h.cursor == 0 {
				p := new(producer)
				p.name = h.name
				p.ticker = h.ticker
				return p, p.Init()
			} else {
				c := new(consumer)
				c.name = h.name
				c.ticker = h.ticker
				return c, c.Init()
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			h.ticker.Flag = NEXT
			return h, nil

		case "ctrl+c", "q":
			return h, tea.Quit
		case "left":
			h.cursor = (h.cursor - 1) % 2
		case "right":
			h.cursor = (h.cursor + 1) % 2
		case "backspace":
			if len(h.name) > 0 {
				h.name = h.name[:len(h.name)-1]
			}

		case "down", "up":

		default:
			h.name += msg.String()
			return h, nil
		}

	}
	return h, nil
}

func (h *hello) View() string {
	one, two := "", ""
	if h.cursor == 0 {
		one = "->"
	} else {
		two = "->"
	}
	return fmt.Sprintf("欢迎您：请输入你的名字，并选择角色：\n%s%s\n%s1.生产者\t%s2.消费者", h.name, h.shane, one, two)
}

func main() {
	h := &hello{}
	cmd := tea.NewProgram(h)
	if err := cmd.Start(); err != nil {
		fmt.Println("start failed:", err)
		os.Exit(1)
	}
}
