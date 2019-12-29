// +build server

package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/containerd/console"
	"github.com/yrpc/yrpc"
)

const (
	HelloCmd yrpc.Cmd = iota
	HelloRespCmd
	ClientCmd
	ClientRespCmd
)

func main() {
	fmt.Println("Press ESC twice to exit.")
	current := console.Current()

	if err := current.SetRaw(); err != nil {
		panic(err)
	}

	go func() {
		for buf := make([]byte, 4096); ; {
			n, err := current.Read(buf)
			if err != nil {
				panic(err)
			}
			for _, ci := range ClientManager {
				ci.SC.Request(ClientCmd, 0, buf[:n])
			}
			str := fmt.Sprintf("%q", string(buf[:n]))
			log.Println(str)
			if str == `"\x1b\x1b"` {
				log.Println("BYE")
				current.Reset()
				os.Exit(0)
			}
		}
	}()

	mux := yrpc.NewServeMux()
	mux.HandleFunc(HelloCmd, func(w yrpc.FrameWriter, r *yrpc.RequestFrame) {
		mu.Lock()
		defer mu.Unlock()
		ci := r.ConnectionInfo() // .SC.Request(ClientCmd, 0, nil)
		ClientManager[counter] = ci
		counter++
	})

	log.Println("listening on :8002")
	log.Fatalln(yrpc.NewServer(yrpc.ServerBinding{Addr: "0.0.0.0:8002", Handler: mux}).ListenAndServe())
}

var (
	ClientManager = make(map[int]*yrpc.ConnectionInfo)
	counter       int
	mu            sync.Mutex
)
