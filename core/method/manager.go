package method

import (
	"QuantumMQ-Core/core"
	"QuantumMQ-Core/internal/container"
	"QuantumMQ-Core/protocol"
)

type Manager interface {
	HasMsgQueue(id core.UniqueID) bool
	CreateMsgQueue(id core.UniqueID)
	GetMsgQueue(id core.UniqueID) *container.LinkedList[*protocol.MQMessage]
}
