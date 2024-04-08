package protocol

import (
	"github.com/bytedance/sonic"
	"github.com/golang/protobuf/proto"
	"testing"
)

var msg = &MQRequest{
	Payload: make([]byte, 4096),
}

func BenchmarkJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b, err := sonic.Marshal(msg)
		if err != nil {
			panic(err)
		}
		err = sonic.Unmarshal(b, new(MQRequest))
		if err != nil {
			panic(err)
		}
	}
}
func BenchmarkProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b, err := proto.Marshal(msg)
		if err != nil {
			panic(err)
		}
		err = proto.Unmarshal(b, new(MQRequest))
		if err != nil {
			panic(err)
		}
	}
}
