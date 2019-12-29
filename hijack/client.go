// +build client

package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/yrpc/yrpc"
)

const (
	Ping yrpc.Cmd = iota
	Pong
)

func main() {
	for range time.Tick(time.Second) {
		err := connect()
		log.Println(err)
	}
}

func connect() error {
	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		return err
	}
	conn.Write([]byte("GET /hijack HTTP/1.1\r\nHost: localhost:8000\r\n\r\n"))

	ioutil.ReadAll(io.LimitReader(conn, 1))

	conf := yrpc.ClientConfig{
		OverlayNetwork: func(addr string, dc yrpc.DialConfig) (net.Conn, error) { return conn, nil },
	}

	yconn, err := yrpc.NewConnection("-", conf, nil)
	if err != nil {
		return err
	}
	log.Println("sending Ping")

	w, resp, err := yconn.StreamRequest(Ping, 0, nil)
	if err != nil {
		return err
	}
	w.StartWrite(Ping)
	w.EndWrite(false)
	frame, err := resp.GetFrame()
	if err != nil {
		return err
	}
	log.Println("Response received", string(frame.Payload))

	for {
		f := <-frame.FrameCh()
		if f == nil {
			log.Println("error: connection closed. Reconnect")
			break
		}
		payload := string(f.Payload)
		cmd := ""
		switch f.Cmd {
		case Ping:
			cmd = "Ping"
			w.StartWrite(Pong)
			w.WriteBytes([]byte(payload))
			w.EndWrite(false)
		default:
			cmd = "unknown cmd"
		}
		log.Println(cmd, payload)
	}
	return nil
}
