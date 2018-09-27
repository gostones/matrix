package server

import (
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

//HTTPServer extends net/http Server and
//adds graceful shutdowns
type HTTPServer struct {
	*http.Server
	listener  net.Listener
	running   chan error
	isRunning bool
	closer    sync.Once
}

//NewHTTPServer creates a new HTTPServer
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		Server:   &http.Server{},
		listener: nil,
		running:  make(chan error, 1),
	}
}

func (h *HTTPServer) GoListenAndServe(addr string, handler http.Handler) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	h.isRunning = true

	//
	r := mux.NewRouter()

	//mux := http.NewServeMux()

	//ping
	r.HandleFunc("/ping", ping)

	//apis

	//tunnel
	r.Handle("/tunnel", handler)

	//web proxy
	r.HandleFunc("/web/{port}", func(res http.ResponseWriter, req *http.Request) {
		port := mux.Vars(req)["port"]
		s := MatrixSession{
			Port:      str2port(port),
			Timestamp: time.Now().UnixNano(),
		}

		SetCookie(res, &s)

		log.Printf("Hanle /web/port: %v %v  url: %v", port, s, req.URL)

		http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
	})

	r.PathPrefix("/").Headers("Connection", "Upgrade", "Upgrade", "websocket").HandlerFunc(NewWsProxyHandler())
	r.PathPrefix("/").HandlerFunc(NewProxyHandler())

	h.Handler = r
	//

	h.listener = l
	go func() {
		h.closeWith(h.Serve(l))
	}()
	return nil
}

func (h *HTTPServer) closeWith(err error) {
	if !h.isRunning {
		return
	}
	h.isRunning = false
	h.running <- err
}

func (h *HTTPServer) Close() error {
	h.closeWith(nil)
	return h.listener.Close()
}

func (h *HTTPServer) Wait() error {
	if !h.isRunning {
		return errors.New("Already closed")
	}
	return <-h.running
}
