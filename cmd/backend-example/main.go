package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	hostPort := flag.String("listen.http", ":4000", "hostPort to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(`Request from: ` + r.RemoteAddr)
		bytes, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("Error reading request: %s\n", err)
		}
		log.Println(string(bytes))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`Hello from backend %s to client %s`, *hostPort, r.RemoteAddr)))
	})

	log.Printf("Listening on %s...\n", *hostPort)
	http.ListenAndServe(*hostPort, mux)
}
