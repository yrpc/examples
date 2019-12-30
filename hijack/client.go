// +build client

package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"crypto/tls"
	"time"

	"github.com/yrpc/yrpc"
)

const (
	Ping yrpc.Cmd = iota
	Pong
)

func main() {
	for range time.Tick(time.Second) {
		err := ConnectAndServe()
		log.Println(err)
	}
}

func ConnectAndServe() error {
	// dial
	// conn, err := tls.Dial("tcp", ":8884")
	conn, err := tls.Dial("tcp", "libredot.com:https", nil)
	if err != nil {
		return err
	}

	// connect
	conn.Write([]byte("GET /api/yrpc HTTP/1.1\r\nHost: localhost:8000\r\n\r\n"))
	// conn.Write([]byte("GET /hijack HTTP/1.1\r\nHost: localhost:8000\r\n\r\n"))

	// connect - confirm
	ioutil.ReadAll(io.LimitReader(conn, 1))

	// dial2
	conf := yrpc.ClientConfig{
		OverlayNetwork: func(addr string, dc yrpc.DialConfig) (net.Conn, error) { return conn, nil },
	}

	yconn, err := yrpc.NewConnection("-", conf, nil)
	if err != nil {
		return err
	}
	log.Println("sending Ping")

	// connect2
	w, resp, err := yconn.StreamRequest(Ping, 0, nil)
	if err != nil {
		return err
	}
	w.StartWrite(Ping)
	// # TODO change this part to sending client info
	// w.WriteBytes(...)
	w.EndWrite(false)

	// connect2 confirm
	frame, err := resp.GetFrame()
	if err != nil {
		return err
	}
	log.Println("Response received", string(frame.Payload))

	// serve
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
