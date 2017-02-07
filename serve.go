package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Static handles serving static files from disk
var Static = http.FileServer(http.Dir("./static"))

// Fallthrough handler: catches all requests that don't match an explicit route
func Fallthrough(w http.ResponseWriter, r *http.Request) {
	path := r.RequestURI

	if strings.HasPrefix(path, "/http://") || strings.HasPrefix(path, "/https://") {
		Proxy(w, r)
		return
	}

	http.NotFound(w, r)
}

// Handle a proxy request
func Proxy(w http.ResponseWriter, r *http.Request) {
	DefaultProxy.ServeHTTP(w, r)
}

func main() {
	router := httprouter.New()
	router.Handler("GET", "/", Static)
	router.Handler("GET", "/assets/*path", Static)
	router.NotFound = http.HandlerFunc(Fallthrough)

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}
