package injecture

import (
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type RewritingTransport struct {
	transport http.RoundTripper
}

var request_header_blacklist = [...]string{
	// Passing CF headers on to via causes "Error 1000: DNS points to
	// prohibited IP" errors at CloudFlare, so we have to strip them out
	// here.
	"CF-Connecting-IP",
	"CF-Ray",
	"CF-Visitor",
	"CF-Ipcountry",

	// Generic HTTP auth headers
	"Authorization",
	"Cookie",

	// h CSRF token header
	"X-Csrf-Token",

	// FIXME: For now just disable gzip responses. In future it would be
	// nice to handle this transparently.
	"Accept-Encoding",
}

var response_header_blacklist = [...]string{
	"Set-Cookie",
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

	// If origin request is http, then we need to rewrite asset urls in response
	if req.URL.Scheme == "http" {
		body, err := RewriteInsecureURLs(res.Body)
		if err != nil {
			log.Fatalf("ERROR: couldn't rewrite insecure URLs %v\n", err)
		}
		res.Body = body
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
	// If the request is a passthrough request (/_/http...) then we need to
	// cut the underscore prefix.
	if strings.HasPrefix(target, "_/") {
		target = target[2:]
	}
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

	for _, header := range request_header_blacklist {
		req.Header.Del(header)
	}

	userAgent := req.Header.Get("User-Agent")
	if userAgent != "" {
		userAgent += " "
	}
	userAgent += "Hypothesis-Via"
	req.Header.Set("User-Agent", userAgent)
}

func RewriteResponse(res *http.Response) error {
	for _, header := range response_header_blacklist {
		res.Header.Del(header)
	}

	location, _ := res.Location()
	if location != nil {
		newLocation := os.Getenv("HEROKU_URL") + "/" + location.String()
		res.Header.Set("Location", newLocation)
		return nil
	}

	return nil
}

var RewriteProxy = &httputil.ReverseProxy{
	Director:       RewriteRequest,
	Transport:      &RewritingTransport{},
	ModifyResponse: RewriteResponse,
}

var PassthroughProxy = &httputil.ReverseProxy{Director: RewriteRequest}