package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/guregu/kami"
)

func GetFoobar(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "foobar")
}

func NewRouter() *kami.Mux {
	mux := kami.New()

	// Routes
	mux.Get("/foobar", GetFoobar)

	return mux
}

func main() {
	location := net.JoinHostPort("0.0.0.0", "8888")
	hdlr := NewRouter()

	apiserver := &http.Server{
		Addr:         location,
		Handler:      hdlr,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}

	apiserver.ListenAndServe()
}
