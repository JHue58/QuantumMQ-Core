package qnet

import (
	connect2 "QuantumMQ-Core/internal/qnet/connect"
	qio2 "QuantumMQ-Core/internal/qnet/qio"
	receive2 "QuantumMQ-Core/internal/qnet/receive"
	"QuantumMQ-Core/internal/qnet/send"
	"QuantumMQ-Core/pkg/errors"
	"QuantumMQ-Core/pkg/goroutine"
	logger2 "QuantumMQ-Core/pkg/logger"
	"net"
)

// NewQuantumClient 新建一个基于普通TCP连接的Client
func NewQuantumClient(address string, sConf send.QTPSenderConfig, wConf qio2.QTPWriterConfig, handler *receive2.CallBackHandler) (*send.QTPSender, *errors.QError) {
	reconnectHandle := func() (conn net.Conn, err *errors.QError) {
		return GetDial(address)
	}
	conn, err := reconnectHandle()
	if err != nil {
		return nil, err
	}

	logger2.GetLogConfig("QTPClient").Level = logger2.WarnLevel
	qLogger := logger2.GetQLogger("QTPClient")

	gm := &goroutine.GoManager{}
	rc := connect2.NewReconnecter(sConf.ReconnectAttempts, reconnectHandle)
	qConn := connect2.NewQTPConn(conn, rc)

	sender := send.NewSender(wConf.SendCap)
	receiver := receive2.NewQTPReceiver()
	writer := qio2.NewQTPWriter(qConn, wConf)
	reader := qio2.NewQTPReader(qConn, qio2.QTPReaderConfigDefault())

	// 设置conn
	sender.SetQTPConn(qConn)
	receiver.SetQTPConn(qConn)

	// 设置logger
	sender.SetLogger(qLogger)
	rc.SetLogger(qLogger)
	receiver.SetLogger(qLogger)

	// 设置gm
	sender.SetGoManager(gm)
	writer.SetGoManager(gm)

	// 设置CallBacker
	receiver.SetCallBacker(&receive2.CallBacker{
		Handler: handler,
		Sender:  sender,
	})

	// 设置reader和writer
	sender.SetQTPReader(reader)
	sender.SetQTPWriter(writer)
	receiver.SetQTPReader(reader)
	receiver.SetQTPWriter(writer)

	// 启动！
	reader.Start()
	receiver.Start()
	writer.Start()
	sender.Start()

	return sender, nil
}
