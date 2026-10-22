// A basic HTTP server.
// By default, it serves the current working directory on port 8080.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

var (
	listen = flag.String("listen", ":8080", "listen address")
	dir    = flag.String("dir", ".", "directory to serve")
)

func main() {
	flag.Parse()
	log.Printf("listening on %q...", *listen)
	// ReadHeaderTimeout: without it a client that never finishes its headers
	// holds a connection open for ever (gosec G114, slowloris).
	srv := &http.Server{
		Addr:              *listen,
		Handler:           http.FileServer(http.Dir(*dir)),
		ReadHeaderTimeout: 10 * time.Second,
	}
	err := srv.ListenAndServe()
	log.Fatalln(err)
}
