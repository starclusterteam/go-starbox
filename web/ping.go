package web

import "net/http"

func ping(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"ping": "pong"}
	WriteJSON(w, http.StatusOK, data)
}

// Ping is the ping HandlerFunc
var Ping = http.HandlerFunc(ping)
