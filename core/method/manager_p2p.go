package method

import (
	"QuantumMQ-Core/core"
	"QuantumMQ-Core/internal/container"
	"QuantumMQ-Core/protocol"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type P2PManager struct {
	msgQueueMap cmap.ConcurrentMap[core.UniqueID, *container.LinkedList[*protocol.MQMessage]]
}

func NewP2PManager() *P2PManager {
	mg := new(P2PManager)
	mg.msgQueueMap = cmap.NewStringer[core.UniqueID, *container.LinkedList[*protocol.MQMessage]]()
	return mg
}

func (manager *P2PManager) HasMsgQueue(id core.UniqueID) bool {
	ok := manager.msgQueueMap.Has(id)
	return ok
}
func (manager *P2PManager) CreateMsgQueue(id core.UniqueID) {
	if manager.HasMsgQueue(id) {
		return
	}
	msgQueue := container.NewLinkedList[*protocol.MQMessage]()
	manager.msgQueueMap.Set(id, msgQueue)
}

func (manager *P2PManager) DeleteMsgQueue(id core.UniqueID) {
	manager.msgQueueMap.Remove(id)
}

func (manager *P2PManager) GetMsgQueue(id core.UniqueID) *container.LinkedList[*protocol.MQMessage] {
	msgQueue, ok := manager.msgQueueMap.Get(id)
	if !ok {
		msgQueue = container.NewLinkedList[*protocol.MQMessage]()
		manager.msgQueueMap.Set(id, msgQueue)
	}
	return msgQueue
}

func (manager *P2PManager) GetAllQueue() []*container.LinkedList[*protocol.MQMessage] {
	keys := manager.msgQueueMap.Keys()
	if len(keys) == 0 {
		return nil
	}
	ls := make([]*container.LinkedList[*protocol.MQMessage], len(keys))
	for i := 0; i < len(keys); i++ {
		ls[i], _ = manager.msgQueueMap.Get(keys[i])
	}
	return ls
}
