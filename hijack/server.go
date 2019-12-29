package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/btwiuse/conntroll/pkg/wrap"
	"github.com/yrpc/yrpc"
)

// lys implements net.Listener
type lys struct {
	conns chan net.Conn
}

func (l *lys) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("hijacking /hijack")
		conn, _ := wrap.Hijack(w)
		log.Println("hijacked /hijack, sending conn to l.conns")
		l.conns <- conn
		conn.Write([]byte(" "))
		log.Println("sent")
	})
}

func (l *lys) SetDeadline(t time.Time) error {
	return nil
}

func (l *lys) Accept() (net.Conn, error) {
	return (<-l.conns).(*wrap.Conn).TCPConn, nil
}

func (l *lys) Close() error {
	return nil
}

func (l *lys) Addr() net.Addr {
	return l
}

func (l *lys) Network() string {
	return "hijack"
}

func (l *lys) String() string {
	return l.Network()
}

func NewLys() *lys {
	return &lys{
		conns: make(chan net.Conn, 100),
	}
}

const (
	Ping yrpc.Cmd = iota
	Pong
)

func main() {
	ly := NewLys()
	mux := http.NewServeMux()
	mux.Handle("/hijack", ly.Handler())

	ymux := yrpc.NewServeMux()
	ymux.Handle(Ping, yrpc.HandlerFunc(func(w yrpc.FrameWriter, r *yrpc.RequestFrame) {
		w.StartWrite(r.RequestID, Pong, yrpc.StreamFlag)
		w.WriteBytes([]byte("stream begins"))
		w.EndWrite()

		// consule the first resp
		<-r.FrameCh()

		quit := make(chan struct{})
		go func() {
			for {
				f := <-r.FrameCh()
				if f == nil {
					log.Println("nil chan")
					close(quit)
					return
				}
				log.Println("client response:", string(f.Payload))
			}
		}()

		for i := 0; ; i++ {
			select {
			case <-quit:
				break
			default:
				payload := fmt.Sprintf("server request %d", i)
				w.StartWrite(r.RequestID, Ping, yrpc.StreamFlag)
				w.WriteBytes([]byte(payload))
				w.EndWrite() // send frame
				log.Println(payload)
			}
			time.Sleep(time.Second)
		}
	}))

	ys := yrpc.NewServer(yrpc.ServerConfig{
		ListenFunc: func(network, addr string) (net.Listener, error) { return ly, nil },
		Handler:    ymux,
	})
	go func() {
		ys.ListenAndServe()
	}()

	http.ListenAndServe(":8000", mux)
}
