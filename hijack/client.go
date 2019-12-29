// +build client

package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/yrpc/yrpc"
)

const (
	Ping yrpc.Cmd = iota
	Pong
)

func main() {
	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("GET /hijack HTTP/1.1\r\nHost: localhost:8000\r\n\r\n"))

	ioutil.ReadAll(io.LimitReader(conn, 1))

	/*
		handler := yrpc.HandlerFunc(func(w yrpc.FrameWriter, r *yrpc.RequestFrame) {
			log.Println("Handler invoked")
			for {
				f := <-r.FrameCh()
			}
		})
	*/

	conf := yrpc.ClientConfig{
		OverlayNetwork: func(addr string, dc yrpc.DialConfig) (net.Conn, error) { return conn, nil },
		// Handler:        handler,
	}

	yconn, _ := yrpc.NewConnection("-", conf, nil)
	log.Println("sending Ping")

	// _, resp, _ := yconn.Request(Ping, 0, nil)
	w, resp, _ := yconn.StreamRequest(Ping, 0, nil)
	w.StartWrite(Ping)
	w.EndWrite(false)
	frame, _ := resp.GetFrame()
	log.Println("Response received", string(frame.Payload))
	for {
		f := <-frame.FrameCh()
		payload := string(f.Payload)
		cmd := ""
		switch f.Cmd {
		case Ping:
			cmd = "ping"
		default:
			cmd = "unknown cmd"
		}
		log.Println(cmd, payload)
		w.StartWrite(Pong)
		w.WriteBytes([]byte(payload))
		w.EndWrite(false)
	}
	select {}
}
