package protocol

import "sync"

type Recyclable interface {
	Recycle()
}

var reqPool = sync.Pool{New: func() any {
	return new(MQRequest)
}}

var respPool = sync.Pool{New: func() any {
	return new(MQResponse)
}}

var multiRespPool = sync.Pool{New: func() any {
	return new(MQMultiResponse)
}}

var msgPool = sync.Pool{New: func() any {
	return new(MQMessage)
}}

func NewMQRequest() *MQRequest {
	return reqPool.Get().(*MQRequest)
}

func NewMQResponse() *MQResponse {
	return respPool.Get().(*MQResponse)
}

func NewMQMultiResponse() *MQMultiResponse {
	return multiRespPool.Get().(*MQMultiResponse)
}

func NewMQMessage() *MQMessage {
	return msgPool.Get().(*MQMessage)
}

func (x *MQRequest) Recycle() {
	x.Receiver = ""
	x.Sender = ""
	x.Payload = nil
	reqPool.Put(x)
}

func (x *MQResponse) Recycle() {
	if x.Message != nil {
		x.Message.Recycle()
	}

	x.Message = nil
	x.Remaining = 0
	x.Error = ""
	respPool.Put(x)
}

func (x *MQMultiResponse) Recycle() {
	if x.Responses != nil {
		for i := 0; i < len(x.Responses); i++ {
			x.Responses[i].Recycle()
		}
	}
	x.Responses = nil
	x.Error = ""
	x.Remaining = 0

	multiRespPool.Put(x)
}

func (x *MQMessage) Recycle() {
	//x.Source = ""
	//x.Payload = nil
	//
	//msgPool.Put(x)
}
