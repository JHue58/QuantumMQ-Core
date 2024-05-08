package main

import (
	"QuantumMQ-Core/internal/qnet"
	"QuantumMQ-Core/internal/qnet/qc"
	"QuantumMQ-Core/internal/qnet/qio"
	"QuantumMQ-Core/internal/qnet/receive"
	"QuantumMQ-Core/internal/qnet/send"
	"QuantumMQ-Core/pkg/errors"
	"QuantumMQ-Core/protocol"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/golang/protobuf/proto"
	"math"
)

type consumer struct {
	name   string
	cli    *send.QTPSender
	ticker *TeaTicker

	msgs  string
	shane string

	cursorLR int
	itemsLR  []string

	topic string

	pulls int
}

func (c *consumer) Init() tea.Cmd {
	c.ticker.Flag = NULL
	h := receive.NewCallBackHandler()
	h.SyncACK(func(sender *send.QTPSender, data *qc.QTPData) {

		if data.ParserError != nil {
			c.msgs += fmt.Sprintf("[MSG] %s\n", data.ParserError.Error())
			return
		}
		resp := protocol.NewMQResponse()
		_ = proto.Unmarshal(data.Data, resp)
		if resp.Error != "" {
			c.msgs += fmt.Sprintf("[MSG] %s\n", resp.Error)
			return
		}
		c.msgs += fmt.Sprintf("[MSG] %s\t来自\t%s\t剩余%d\n", string(resp.Message.Payload), resp.Message.Source, resp.Remaining)
	})
	cli, err := qnet.NewQuantumClient("159.75.93.21:8888", send.QTPSenderConfigDefault(), qio.QTPWriterConfigDefault(), h)
	if err != nil {
		fmt.Println(err.Error())
		return tea.Quit
	}
	c.cli = cli
	c.shane = "_"
	c.pulls = 1
	return c.ticker.Update()
}

func (c *consumer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *TeaTicker:
		if msg.Flag == NULL {
			if c.shane == " " {
				c.shane = "_"
			} else {
				c.shane = " "
			}
			return c, msg.Update()
		} else {
			h := new(hello)
			h.name = c.name
			return h, h.Init()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			c.ticker.Flag = NEXT
			c.cli.Close()
			return c, nil
		case "left":
			c.cursorLR = int(math.Abs(float64((c.cursorLR - 1) % 2)))
		case "right":
			c.cursorLR = int(math.Abs(float64((c.cursorLR + 1) % 2)))
		case "backspace":
			if c.cursorLR == 1 && len(c.topic) > 0 {
				c.topic = c.topic[:len(c.topic)-1]
			}
		case "enter":
			req := protocol.NewMQRequest()
			req.Sender = c.name
			req.Receiver = c.topic
			req.Action = protocol.Action_Pull
			req.Pulls = uint64(c.pulls)
			if c.cursorLR == 0 {
				req.Mode = protocol.Mode_P2P
			} else {
				req.Mode = protocol.Mode_PubSub
			}
			b, _ := proto.Marshal(req)
			c.cli.SendSyncACK(b, qc.BINARY, func(seq uint64, err *errors.QError) {
				if err != nil {
					c.msgs += fmt.Sprintf("SyncACK: %s", err.Error())
					return
				}

			})
		case "up", "down":
		default:
			if c.cursorLR == 1 {
				c.topic += msg.String()
			}
		}
	}
	return c, nil
}

func (c *consumer) View() string {
	title := fmt.Sprintf("欢迎你 %s 现在是消费者模式\n", c.name)
	inputs := ""
	if c.cursorLR == 1 {
		inputs = fmt.Sprintf("Topic: %s%s", c.topic, c.shane)
	}
	selects := ""
	if c.cursorLR == 0 {
		selects = "->1.P2P模式\t2.发布订阅模式"
	} else {
		selects = "1.P2P模式\t->2.发布订阅模式"
	}

	return title + "\n" + inputs + "\n\n" + selects + "\n" + c.msgs
}
