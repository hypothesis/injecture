package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"regexp"
)

var (
	SCAN_SIZE  = 8192
	INJECT_RE  = regexp.MustCompile(`(?i)<html[\s>]|<head[\s>]|(<[a-z\/])`)
	INJECT_TPL = template.Must(template.ParseFiles("inject.html"))
)

type TemplateData struct {
	URL string
}

func Inject(url string, src io.ReadCloser) io.ReadCloser {
	buf := make([]byte, SCAN_SIZE)
	_, err := src.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	payload := &bytes.Buffer{}
	INJECT_TPL.Execute(payload, TemplateData{
		URL: url,
	})

	result := inject(buf, payload.Bytes())
	return struct {
		io.Reader
		io.Closer
	}{
		io.MultiReader(
			bytes.NewBuffer(result),
			src,
		),
		src,
	}
}

// var testcases = []string {
// 	"<!doctype html><foo>hello</foo>",
// 	"<html><p>Whoops!</p>",
// 	"<html><head><title>Hello there!</head><body>Whoops</body>",
// 	"<!doctype html><foo>hello</foo>",
// 	"<HTML><p>Whoops!</p>",
// 	"<HTML><HEAD><title>Hello there!</head><body>Whoops</body>",
// }

func inject(input []byte, payload []byte) (output []byte) {
	// We don't care about the matches for <html> or <head>, we care about the
	// third match, which is a tag starting with either a character in [a-z/]
	tag := 2
	// Find all matches
	matches := INJECT_RE.FindAllSubmatchIndex(input, -1)
	for _, match := range matches {
		// If the third group has matched, then inject the payload
		if match[tag] > -1 {
			result := []byte{}
			result = append(result, input[:match[tag]]...)
			result = append(result, payload...)
			result = append(result, input[match[tag]:]...)
			return result
		}
	}
	// We fell through without matching
	return input
}
