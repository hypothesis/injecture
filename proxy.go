package main

import (
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RewritingTransport struct {
	transport http.RoundTripper
}

func (r *RewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.transport == nil {
		r.transport = http.DefaultTransport
	}
	res, err := r.transport.RoundTrip(req)
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

	// OK, inject...
	res.Body = Inject(req.RequestURI[1:], res.Body)

	// Remove the Content-Length header -- we've changed the length and Go's
	// HTTP server will happily serve this as T-E: chunked.
	res.Header.Del("Content-Length")
	return res, nil
}

func RewriteRequest(req *http.Request) {
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

var DefaultProxy = &httputil.ReverseProxy{
	Director:  RewriteRequest,
	Transport: &RewritingTransport{},
}
