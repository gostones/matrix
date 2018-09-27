package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

//type WebProxy struct {
//	path  string
//	mux   *http.ServeMux
//	vport int //vhost_http_port
//}
//
//func NewWebProxy(path string, mux *http.ServeMux) *WebProxy {
//	s := WebProxy{
//		path: path,
//		mux:  mux,
//	}
//	return &s
//}
//
//func (r *WebProxy) Handle() {
//	r.mux.Handle(fmt.Sprintf("%v/websocket/", r.path), r.wsProxy(fmt.Sprintf("ws://localhost:%v/websocket/", r.vport)))
//	r.mux.Handle(fmt.Sprintf("%v/", r.path), http.StripPrefix(r.path, NewProxyHandler()))
//}
//
//func (r *WebProxy) wsProxy(remoteUrl string) http.Handler {
//	target := toUrl(remoteUrl)
//	handler := websocketproxy.NewProxy(target)
//
//	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
//		fmt.Fprintf(os.Stdout, "wsProxy: %v\n", req.URL)
//
//		handler.ServeHTTP(res, req)
//	})
//}

//func (r *WebProxy) httpProxy(remoteUrl string) http.Handler {
//	target := toUrl(remoteUrl)
//	handler := httputil.NewSingleHostReverseProxy(target)
//	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
//		fmt.Fprintf(os.Stdout, "httpProxy: %v\n", req.URL)
//
//		handler.ServeHTTP(res, req)
//	})
//}

func toUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

//
func StripPrefix(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// /p/port/
		if p := strings.SplitN(r.URL.Path, "/", 4); len(p) > 3 {
			fmt.Printf("StripPrefix: %v %v\n", p, r.URL)

			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p[3]
			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	}
}
