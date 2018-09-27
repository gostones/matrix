package server

import (
	"net/http"
)

func ping(w http.ResponseWriter, req *http.Request) {
	type health struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		Timestamp int64  `json:"timestamp"`
	}
	name := "Matrix"
	m := health{Name: name, Status: "UP", Timestamp: CurrentTimestamp()}
	HandleJson(&m, w, req)
}
