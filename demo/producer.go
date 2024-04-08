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

type producer struct {
	ticker *TeaTicker

	name string

	cli      *send.QTPSender
	sendTips string
	shane    string

	cursorLR int
	itemsLR  []string

	cursorUD int
	itemsUD  []string

	inputs [][]string
}

func (p *producer) Init() tea.Cmd {
	p.ticker.Flag = NULL
	h := receive.NewCallBackHandler()
	h.SyncACK(func(sender *send.QTPSender, data *qc.QTPData) {

	})

	cli, err := qnet.NewQuantumClient(":8888", send.QTPSenderConfigDefault(), qio.QTPWriterConfigDefault(), h)
	if err != nil {
		fmt.Println(err.Error())
		return tea.Quit
	}
	p.cli = cli
	p.shane = "_"
	p.itemsLR = []string{
		"P2P", "发布订阅模式",
	}
	p.itemsUD = []string{
		"接收方", "内容",
	}
	p.inputs = [][]string{
		{"", ""}, {"", ""},
	}

	return p.ticker.Update()
}

func (p *producer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *TeaTicker:
		if msg.Flag == NULL {
			if p.shane == " " {
				p.shane = "_"
			} else {
				p.shane = " "
			}
			return p, msg.Update()
		} else {
			h := new(hello)
			h.name = p.name
			return h, h.Init()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
			p.ticker.Flag = NEXT
			p.cli.Close()
			return p, nil
		case "left":
			p.cursorLR = int(math.Abs(float64((p.cursorLR - 1) % 2)))
			if p.cursorLR == 0 {
				p.itemsUD = []string{
					"接收方", "内容",
				}

			} else {
				p.itemsUD = []string{
					"Topic", "内容",
				}

			}
		case "right":
			p.cursorLR = int(math.Abs(float64((p.cursorLR + 1) % 2)))
			if p.cursorLR == 0 {
				p.itemsUD = []string{
					"接收方", "内容",
				}

			} else {
				p.itemsUD = []string{
					"Topic", "内容",
				}

			}
		case "up":
			p.cursorUD = int(math.Abs(float64((p.cursorUD - 1) % 2)))
		case "down":
			p.cursorUD = int(math.Abs(float64((p.cursorUD + 1) % 2)))
		case "backspace":
			if len(p.inputs[p.cursorLR][p.cursorUD]) > 0 {
				p.inputs[p.cursorLR][p.cursorUD] = p.inputs[p.cursorLR][p.cursorUD][:len(p.inputs[p.cursorLR][p.cursorUD])-1]
			}
		case "enter":
			req := protocol.NewMQRequest()

			req.Action = protocol.Action_Push
			req.Sender = p.name
			req.Receiver = p.inputs[p.cursorLR][0]
			req.Payload = []byte(p.inputs[p.cursorLR][1])
			if p.cursorLR == 0 {
				req.Mode = protocol.Mode_P2P
			} else {
				req.Mode = protocol.Mode_PubSub
			}
			b, _ := proto.Marshal(req)
			p.cli.SendSyncACK(b, qc.BINARY, func(seq uint64, err *errors.QError) {
				if err != nil {
					p.sendTips = fmt.Sprintf("SyncACK: %s", err.Error())
					return
				}
				p.sendTips = "发送成功！"
			})
			p.inputs[p.cursorLR][1] = ""

		default:
			p.inputs[p.cursorLR][p.cursorUD] += msg.String()
			return p, nil
		}

	}
	return p, nil
}

func (p *producer) View() string {
	title := fmt.Sprintf("欢迎你 %s 现在是生产者模式\n", p.name)
	inputs := ""
	if p.cursorUD == 0 {
		input1 := fmt.Sprintf("%s: %s%s", p.itemsUD[0], p.inputs[p.cursorLR][0], p.shane)
		input2 := fmt.Sprintf("%s: %s", p.itemsUD[1], p.inputs[p.cursorLR][1])
		inputs = input1 + "\n" + input2
	} else {
		input1 := fmt.Sprintf("%s: %s", p.itemsUD[0], p.inputs[p.cursorLR][0])
		input2 := fmt.Sprintf("%s: %s%s", p.itemsUD[1], p.inputs[p.cursorLR][1], p.shane)
		inputs = input1 + "\n" + input2
	}
	selects := ""
	if p.cursorLR == 0 {
		selects = fmt.Sprintf("->1.%s\t2.%s", p.itemsLR[0], p.itemsLR[1])
	} else {
		selects = fmt.Sprintf("1.%s\t->2.%s", p.itemsLR[0], p.itemsLR[1])
	}
	return title + "\n" + inputs + "\n\n" + selects + "\n" + p.sendTips + "\n"

}
