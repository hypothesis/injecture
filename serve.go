package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hypothesis/injecture/injecture"
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

	if strings.HasPrefix(path, "/_/http://") || strings.HasPrefix(path, "/_/https://") {
		Passthrough(w, r)
		return
	}

	http.NotFound(w, r)
}

// Handle a rewriting proxy request
func Proxy(w http.ResponseWriter, r *http.Request) {
	injecture.RewriteProxy.ServeHTTP(w, r)
}

// Handle a passthrough proxy request
func Passthrough(w http.ResponseWriter, r *http.Request) {
	injecture.PassthroughProxy.ServeHTTP(w, r)
}

func main() {
	fmt.Println("Config APP_URL: ", os.Getenv("APP_URL"))
	fmt.Println("Config PORT: ", os.Getenv("PORT"))

	router := httprouter.New()
	router.Handler("GET", "/", Static)
	router.Handler("GET", "/assets/*path", Static)
	router.NotFound = http.HandlerFunc(Fallthrough)

	log.Fatal(http.ListenAndServe("0.0.0.0:"+os.Getenv("PORT"), router))
}
