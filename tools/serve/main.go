// serve previews the site/ directory locally the way the static host
// serves it: /api/topics rewrites to /api/topics.json and API replies
// are JSON. Usage: go run ./tools/serve [-addr :8391] [-dir site]
package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8391", "listen address")
	dir := flag.String("dir", "site", "site directory")
	flag.Parse()

	fs := http.FileServer(http.Dir(*dir))
	log.Printf("serving %s on http://%s", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/topics" {
			r.URL.Path = "/api/topics.json"
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		fs.ServeHTTP(w, r)
	})))
}
