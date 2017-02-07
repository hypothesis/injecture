package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var INJECT_TPL = template.Must(template.New("inject").Parse(`
	<base href="{{.URI}}">
	<link rel="canonical" href="{{.URI}}">
	<script async src="//hypothes.is/embed.js"></script>
`))

type TemplateData struct {
	URI string
}

type InjectorTransport struct {
	dfl http.RoundTripper
}

func (it *InjectorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if it.dfl == nil {
		it.dfl = http.DefaultTransport
	}
	res, err := it.dfl.RoundTrip(req)
	if err != nil {
		return res, err
	}

	header := res.Header.Get("Content-Type")
	mtype, _, err := mime.ParseMediaType(header)
	// If we can't parse the media type, we can't make any assumptions, so just
	// pass the request through.
	if err != nil {
		return res, nil
	}

	if mtype != "text/html" {
		return res, nil
	}

	// Remove the Content-Length header -- we've changed the length and Go's
	// HTTP server will happily serve this as T-E: chunked.
	res.Header.Del("Content-Length")

	payload := &bytes.Buffer{}
	INJECT_TPL.Execute(payload, TemplateData{
		URI: req.RequestURI[1:],
	})
	fmt.Printf("---> %s\n", req.RequestURI[1:])
	res.Body = Inject(res.Body, payload.Bytes())
	return res, nil
}

func NewInjectingProxy() http.Handler {
	director := func(req *http.Request) {
		// The requested URL is assumed to be everything following the initial
		// slash of the path.
		target := req.RequestURI[1:]
		// If we can't parse the request as a URL, then we won't try and proxy it.
		targetURL, err := url.ParseRequestURI(target)
		if err != nil {
			log.Fatalf("ERROR: couldn't parse URL %v\n", err)
		}
		req.URL.Scheme = targetURL.Scheme
		req.Host = targetURL.Host
		req.URL.Host = targetURL.Host
		req.URL.Path = targetURL.Path
		req.URL.RawQuery = targetURL.RawQuery

		// FIXME: For now just disable gzip responses. In future it would be
		// nice to handle this transparently.
		req.Header.Del("Accept-Encoding")
	}
	return &httputil.ReverseProxy{Director: director, Transport: &InjectorTransport{}}
}

type Mux struct {
	proxy http.Handler
}

func NewMux() *Mux {
	return &Mux{
		proxy: NewInjectingProxy(),
	}
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := url.ParseRequestURI(r.RequestURI[1:])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	m.proxy.ServeHTTP(w, r)
}

func main() {
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", NewMux()))
}
