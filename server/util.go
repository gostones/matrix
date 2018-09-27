package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var ContentType = struct {
	JSON string
	HTML string
	JS   string
	CSS  string
	BIN  string
}{
	JSON: "application/json",
	HTML: "text/html",
	JS:   "application/javascript",
	CSS:  "text/css",
	BIN:  "application/octet-stream",
}

func HandleJson(m interface{}, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			res.WriteHeader(http.StatusInternalServerError)
			//log.Errorf("Handle: %s", r)
		}
	}()

	res.Header().Set("Content-Type", ContentType.JSON)
	res.WriteHeader(http.StatusOK)

	b, _ := json.Marshal(m)
	fmt.Fprintf(res, string(b))
}

func CurrentTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
