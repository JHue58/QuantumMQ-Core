package server

import (
	"QuantumMQ-Core/core"
	"QuantumMQ-Core/core/method"
	"QuantumMQ-Core/internal/container"
	"QuantumMQ-Core/internal/qnet"
	"QuantumMQ-Core/internal/qnet/qc"
	"QuantumMQ-Core/internal/qnet/qio"
	"QuantumMQ-Core/internal/qnet/receive"
	"QuantumMQ-Core/internal/qnet/send"
	"QuantumMQ-Core/pkg/errors"
	"QuantumMQ-Core/protocol"
	"github.com/golang/protobuf/proto"
)

type MQServer struct {
	psManager  *method.PubSubManager
	p2pManager *method.P2PManager
	svr        *qnet.QTPServer
}

func NewMQServer() *MQServer {
	s := &MQServer{
		psManager:  method.NewPubSubManager(),
		p2pManager: method.NewP2PManager(),
	}
	return s
}

func (s *MQServer) Run(address string) error {
	cb := receive.NewCallBackHandler()
	cb.NoACK(s.noACK)
	cb.AsyncACK(s.asyncACK)
	cb.SyncACK(s.syncACK)
	svr, err := qnet.NewQuantumServer(address, cb, qio.QTPWriterConfigDefault())
	if err != nil {
		return err
	}
	svr.Start()
	return nil
}

func (s *MQServer) noACK(sender *send.QTPSender, data *qc.QTPData) {
	req := s.parseReq(sender, data)
	if req == nil {
		return
	}
	switch req.Action {
	case protocol.Action_Push:
		s.onPushAction(req)
	case protocol.Action_Pull:
		var rp protocol.Recyclable
		link := s.onPullAction(req)
		var d []byte
		if req.Pulls <= 1 {
			resp := protocol.NewMQResponse()
			next, err := link.Next()
			if err != nil {
				resp.Error = err.Error()
			} else {
				resp.Message = next
				resp.Remaining = uint64(link.Len())
			}
			d, _ = proto.Marshal(resp)
			rp = resp
		} else {
			resp := protocol.NewMQMultiResponse()
			for i := uint64(0); i < req.Pulls; i++ {
				next, err := link.Next()
				if err != nil {
					break
				}
				r := protocol.NewMQResponse()
				r.Message = next
				resp.Responses = append(resp.Responses, r)
			}
			if resp.Responses == nil || len(resp.Responses) == 0 {
				resp.Error = "empty"
			} else {
				resp.Remaining = uint64(link.Len())
			}
			d, _ = proto.Marshal(resp)
			rp = resp

		}

		sender.SendNoACK(d, qc.BINARY, func(seq uint64, err *errors.QError) {
		})
		req.Recycle()
		rp.Recycle()
		link.Release()

	}
}

func (s *MQServer) asyncACK(sender *send.QTPSender, data *qc.QTPData) {
	req := s.parseReq(sender, data)
	if req == nil {
		return
	}
	switch req.Action {
	case protocol.Action_Push:
		s.onPushAction(req)
	case protocol.Action_Pull:
		var rp protocol.Recyclable
		l := s.onPullAction(req)
		link := l.Pin()

		var d []byte
		if req.Pulls <= 1 {
			resp := protocol.NewMQResponse()
			next, err := link.PinNext()
			if err != nil {
				resp.Error = err.Error()
			} else {
				resp.Message = next
				resp.Remaining = uint64(link.Len())
			}
			d, _ = proto.Marshal(resp)
			rp = resp
		} else {
			resp := protocol.NewMQMultiResponse()
			for i := uint64(0); i < req.Pulls; i++ {
				next, err := link.PinNext()
				if err != nil {
					break
				}
				r := protocol.NewMQResponse()
				r.Message = next
				resp.Responses = append(resp.Responses, r)
			}
			if resp.Responses == nil || len(resp.Responses) == 0 {
				resp.Error = "empty"
			} else {
				resp.Remaining = uint64(link.Len())
			}
			d, _ = proto.Marshal(resp)
			rp = resp

		}
		sender.SendAsyncACK(d, qc.BINARY, func(seq uint64, err *errors.QError) {
			if err == nil {
				link.PinRelease()
			} else {
				link.PinRollBack()
			}
		})
		req.Recycle()
		rp.Recycle()

	}
}

func (s *MQServer) syncACK(sender *send.QTPSender, data *qc.QTPData) {

	req := s.parseReq(sender, data)
	if req == nil {
		return
	}
	switch req.Action {
	case protocol.Action_Push:
		s.onPushAction(req)
	case protocol.Action_Pull:
		var rp protocol.Recyclable
		link := s.onPullAction(req)
		var d []byte
		if req.Pulls <= 1 {
			resp := protocol.NewMQResponse()
			next, err := link.Next()
			if err != nil {
				resp.Error = err.Error()
			} else {
				resp.Message = next
				resp.Remaining = uint64(link.Len())
			}
			d, _ = proto.Marshal(resp)
			rp = resp
		} else {
			resp := protocol.NewMQMultiResponse()
			for i := uint64(0); i < req.Pulls; i++ {
				next, err := link.Next()
				if err != nil {
					break
				}
				r := protocol.NewMQResponse()
				r.Message = next
				resp.Responses = append(resp.Responses, r)
			}
			if resp.Responses == nil || len(resp.Responses) == 0 {
				resp.Error = "empty"
			} else {
				resp.Remaining = uint64(link.Len())
			}
			d, _ = proto.Marshal(resp)
			rp = resp

		}
		sender.SendSyncACK(d, qc.BINARY, func(seq uint64, err *errors.QError) {
			if err == nil {
				link.Release()
			} else {
				link.RollBack()
			}
		})
		req.Recycle()
		rp.Recycle()
	}
}

func (s *MQServer) onPushAction(req *protocol.MQRequest) {
	py := make([]byte, len(req.Payload))
	copy(py, req.Payload)
	msg := protocol.NewMQMessage()
	msg.Source = req.Sender
	msg.Payload = py

	switch req.Mode {
	case protocol.Mode_P2P:
		s.p2pManager.GetMsgQueue(core.UniqueID(req.Receiver)).Append(msg)
	case protocol.Mode_PubSub:
		s.psManager.Publish(req)
	}
}

func (s *MQServer) onPullAction(req *protocol.MQRequest) *container.LinkedList[*protocol.MQMessage] {
	switch req.Mode {
	case protocol.Mode_P2P:
		link := s.p2pManager.GetMsgQueue(core.UniqueID(req.Sender))
		return link
	case protocol.Mode_PubSub:
		topicAndUsr := core.UniqueID(req.Receiver).Join(core.UniqueID(req.Sender))
		link := s.psManager.GetMsgQueue(topicAndUsr)
		return link
	}
	return nil
}

func (s *MQServer) parseReq(sender *send.QTPSender, data *qc.QTPData) *protocol.MQRequest {
	if data.ParserError != nil {
		sender.GetLogger().QError(data.ParserError)
		sender.Close()
		return nil
	}
	req := protocol.NewMQRequest()
	err := proto.Unmarshal(data.Data, req)
	if err != nil {
		sender.GetLogger().Error(err.Error())
		sender.Close()
		return nil
	}
	return req
}
