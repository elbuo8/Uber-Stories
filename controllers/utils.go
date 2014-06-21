package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response map[string]interface{}

func (r *Response) String() (s string) {
	b, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(b)
}

func ISR(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func BR(w http.ResponseWriter, r *http.Request, msg error, code int) {
	ServeJSON(w, r, &Response{"error": msg.Error()}, code)
}

func NotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}

func ServeJSON(w http.ResponseWriter, r *http.Request, json *Response, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprint(w, json)
}
