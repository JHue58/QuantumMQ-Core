package server

import (
	"QuantumMQ-Core/internal/qnet"
	"QuantumMQ-Core/internal/qnet/qc"
	"QuantumMQ-Core/internal/qnet/qio"
	"QuantumMQ-Core/internal/qnet/receive"
	"QuantumMQ-Core/internal/qnet/send"
	"QuantumMQ-Core/pkg/errors"
	"QuantumMQ-Core/protocol"
	"fmt"
	"github.com/golang/protobuf/proto"
	"strconv"
	"sync"
	"testing"
)

func TestSvr(t *testing.T) {
	svr := NewMQServer()
	svr.Run(":8888")
}

func BenchmarkCli(b *testing.B) {
	h := receive.NewCallBackHandler()
	h.SyncACK(func(sender *send.QTPSender, data *qc.QTPData) {
	})
	writeConf := qio.QTPWriterConfigDefault()
	writeConf.SendCap = 4096000
	cli, err := qnet.NewQuantumClient(":8888", send.QTPSenderConfigDefault(), writeConf, h)
	if err != nil {
		panic(err)
	}
	req := protocol.NewMQRequest()
	req.Payload = make([]byte, 128)
	req.Action = protocol.Action_Push
	req.Mode = protocol.Mode_P2P
	data, _ := proto.Marshal(req)
	cb := func(seq uint64, err *errors.QError) {
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < b.N; i++ {
		cli.SendNoACK(data, qc.BINARY, cb)
	}
}

func TestCli(t *testing.T) {
	wg := sync.WaitGroup{}
	count := 10

	h := receive.NewCallBackHandler()
	h.SyncACK(func(sender *send.QTPSender, data *qc.QTPData) {
		defer wg.Done()
		if data.ParserError == nil {
			m := protocol.NewMQMultiResponse()
			err := proto.Unmarshal(data.Data, m)
			if err != nil {
				panic(err)
			}
			fmt.Println(m)
		}

	})

	cli, err := qnet.NewQuantumClient(":8888", send.QTPSenderConfigDefault(), qio.QTPWriterConfigDefault(), h)
	if err != nil {
		panic(err)
	}
	//f := make(chan struct{})
	req := protocol.NewMQRequest()

	req.Receiver = "you"
	req.Sender = "me"
	for i := 0; i < count; i++ {
		req.Payload = []byte("this is a payload " + strconv.Itoa(i))
		req.Action = protocol.Action_Push
		req.Mode = protocol.Mode_PubSub
		data, rr := proto.Marshal(req)
		if rr != nil {
			panic(rr)
		}

		cli.SendSyncACK(data, qc.BINARY, func(seq uint64, err *errors.QError) {
			if err != nil {
				panic(err)
			}
		})

	}

	req.Action = protocol.Action_Pull
	req.Sender = "me"
	req.Pulls = uint64(10)
	data, rr := proto.Marshal(req)
	if rr != nil {
		panic(rr)
	}
	cli.SendSyncACK(data, qc.BINARY, func(seq uint64, err *errors.QError) {
		wg.Add(1)
	})
	wg.Wait()
	cli.Close()
}
