// +build client

package main

import (
	"fmt"
	"log"

	"github.com/yrpc/yrpc"
)

const (
	HelloRequest yrpc.Cmd = iota
	HelloRespCmd
	ClientRequest
	ClientRespCmd
)

func main() {
	conf := yrpc.ConnectionConfig{
		Handler: yrpc.HandlerFunc(func(w yrpc.FrameWriter, r *yrpc.RequestFrame) {
			payload := string(r.Payload)
			cmd := ""
			switch r.Cmd {
			case HelloRequest:
				cmd = "hello request"
			case ClientRequest:
				cmd = "client request"
			default:
				cmd = "unknown cmd"
			}
			log.Println(cmd, "received with payload:", payload)
			/*
			w.StartWrite(frame.RequestID, ClientRespCmd, 0)
			w.WriteBytes([]byte("client resp"))
			w.EndWrite()
			*/
		}),
	}
	conn, _ := yrpc.NewConnection("0.0.0.0:8002", conf, func(conn *yrpc.Connection, frame *yrpc.Frame) {
		fmt.Println(string(frame.Payload))
	})
	conn.Request(HelloRequest, yrpc.NBFlag, []byte("xu "))
	select{}
}
