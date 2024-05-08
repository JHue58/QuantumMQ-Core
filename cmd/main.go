package main

import (
	"QuantumMQ-Core/server"
	"log"
)

func main() {
	address := ":8888"
	svr := server.NewMQServer()
	log.Printf("Svr running at %s ...", address)
	err := svr.Run(address)
	if err != nil {
		panic(err)
	}
}
