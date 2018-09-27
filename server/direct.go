package server

import (
	"fmt"
	"github.com/koding/websocketproxy"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

const proxyETagPrefix = "matrix"

func NewProxyHandler() http.HandlerFunc {
	director := func(req *http.Request) {
		// matrix-port-timestamp
		//tag := strings.Split(req.Header.Get("ETAG"), "-")
		//log.Printf("NewProxyHandler director etag: %v url: %v", tag, req.URL)
		//
		//if len(tag) != 3 {
		//	return
		//}

		s := MatrixSession{}
		err := GetCookie(req, &s)
		if err != nil {
			log.Printf("director error: %v", err)
			return
		}
		req.URL.Scheme = "http"
		req.URL.Host = fmt.Sprintf("localhost:%v", s.Port)

		//
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	proxy := &httputil.ReverseProxy{Director: director}

	return func(res http.ResponseWriter, req *http.Request) {
		//etag
		//tag := strings.Split(req.Header.Get("ETAG"), "-")
		s := MatrixSession{}
		err := GetCookie(req, &s)

		if err != nil {
			message := fmt.Sprintf("%v", time.Now().Nanosecond())
			res.Write([]byte(message))
			return
		}
		//ntag := fmt.Sprintf("%v-%v-%v", proxyETagPrefix, s.Port, time.Now().Nanosecond())
		//res.Header().Set("ETag", ntag)

		//SetCookie(res, s)

		log.Printf("NewProxyHandler proxy: %v url: %v", s, req.URL)

		proxy.ServeHTTP(res, req)
	}
}

func NewWsProxyHandler() http.HandlerFunc {

	backend := func(r *http.Request) *url.URL {
		s := MatrixSession{}
		err := GetCookie(r, &s)
		if err != nil {
			log.Printf("backen error: %v", err)
			return r.URL
		}

		// Shallow copy
		u := *r.URL

		u.Scheme = "ws"
		u.Host = fmt.Sprintf("localhost:%v", s.Port)

		u.Fragment = r.URL.Fragment
		u.Path = r.URL.Path
		u.RawQuery = r.URL.RawQuery
		return &u
	}

	proxy := &websocketproxy.WebsocketProxy{Backend: backend}

	return func(res http.ResponseWriter, req *http.Request) {
		//etag
		//tag := strings.Split(req.Header.Get("ETAG"), "-")
		s := MatrixSession{}
		err := GetCookie(req, &s)

		if err != nil {
			message := fmt.Sprintf("%v", time.Now().UnixNano())
			res.Write([]byte(message))
			return
		}
		//ntag := fmt.Sprintf("%v-%v-%v", proxyETagPrefix, s.Port, time.Now().Nanosecond())
		//res.Header().Set("ETag", ntag)

		//SetCookie(res, s)

		log.Printf("NewWsProxyHandler proxy: %v url: %v", s, req.URL)

		proxy.ServeHTTP(res, req)
	}
}

func str2port(s string) int {
	if p, err := strconv.Atoi(s); err == nil {
		return p
	}
	return -1
}
