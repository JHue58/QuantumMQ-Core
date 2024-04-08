package method

import (
	"QuantumMQ-Core/core"
	"QuantumMQ-Core/internal/container"
	"QuantumMQ-Core/protocol"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type PubSubManager struct {
	m cmap.ConcurrentMap[core.UniqueID, *P2PManager]
}

func NewPubSubManager() *PubSubManager {
	mg := new(PubSubManager)
	mp := cmap.NewStringer[core.UniqueID, *P2PManager]()
	mg.m = mp
	return mg
}

func (p *PubSubManager) HasMsgQueue(id core.UniqueID) bool {
	topic, usr := id.Split()
	manager, ok := p.m.Get(topic)
	if !ok {
		return false
	}
	return manager.HasMsgQueue(usr)
}

func (p *PubSubManager) CreateMsgQueue(id core.UniqueID) {
	topic, usr := id.Split()
	manager, ok := p.m.Get(topic)
	if !ok {
		p.m.Set(topic, NewP2PManager())
	}
	manager.CreateMsgQueue(usr)
}

func (p *PubSubManager) GetMsgQueue(id core.UniqueID) *container.LinkedList[*protocol.MQMessage] {
	topic, usr := id.Split()
	manager, ok := p.m.Get(topic)
	if !ok {
		manager = NewP2PManager()
		p.m.Set(topic, manager)
	}
	return manager.GetMsgQueue(usr)
}

func (p *PubSubManager) Subscribe(id core.UniqueID) {
	topic, usr := id.Split()
	manager, ok := p.m.Get(topic)
	if !ok {
		manager = NewP2PManager()
		p.m.Set(topic, manager)
	}
	manager.CreateMsgQueue(usr)
}

func (p *PubSubManager) UnSubscribe(id core.UniqueID) {
	topic, usr := id.Split()
	manager, ok := p.m.Get(topic)
	if !ok {
		return
	}
	manager.DeleteMsgQueue(usr)
}

func (p *PubSubManager) Publish(req *protocol.MQRequest) {
	if req.Action != protocol.Action_Push {
		return
	}
	if req.Mode != protocol.Mode_PubSub {
		return
	}
	py := make([]byte, len(req.Payload))
	copy(py, req.Payload)
	msg := &protocol.MQMessage{
		Source:  req.Sender,
		Payload: py,
	}
	topic := req.Receiver
	manager, ok := p.m.Get(core.UniqueID(topic))
	if !ok {
		manager = NewP2PManager()
		p.m.Set(core.UniqueID(topic), manager)
	}
	manager.CreateMsgQueue(core.UniqueID(req.Sender))
	al := manager.GetAllQueue()
	for i := 0; i < len(al); i++ {
		al[i].Append(msg)
	}
}
