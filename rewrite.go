package main

import (
	"bytes"
	"io"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type BufferCloser struct {
	*bytes.Buffer
}

func (bc *BufferCloser) Close() error {
	return nil
}

func RewriteInsecureURLs(src io.ReadCloser) (io.ReadCloser, error) {
	buffer := &BufferCloser{bytes.NewBufferString("")}

	tokenizer := html.NewTokenizer(src)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		switch tokenType {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				return buffer, nil
			}
			return nil, err

		case html.StartTagToken, html.SelfClosingTagToken:
			switch token.Data {
			case "link":
				rewriteLinkStylesheet(&token)
			case "script":
				rewriteScriptSrc(&token)
			}

			buffer.WriteString(token.String())

		case html.EndTagToken, html.CommentToken, html.DoctypeToken:
			buffer.WriteString(token.String())
		case html.TextToken:
			// token.String() escapes certain characters like quotes,
			// token.Data holds the unescaped version
			buffer.WriteString(token.Data)
		}
	}

	return buffer, nil
}

func rewriteLinkStylesheet(token *html.Token) {
	var href html.Attribute
	isStylesheet := false

	rewrittenAttr := []html.Attribute{}

	for _, attr := range token.Attr {
		switch {
		case attr.Key == "href":
			href = attr
		case attr.Key == "rel" && attr.Val == "stylesheet":
			isStylesheet = true
			rewrittenAttr = append(rewrittenAttr, attr)
		default:
			rewrittenAttr = append(rewrittenAttr, attr)
		}
	}

	if isStylesheet && href.Val != "" {
		val := href.Val
		if strings.HasPrefix(val, "//") {
			val = "https:" + val
		}
		val = os.Getenv("HEROKU_URL") + "/" + val

		href.Val = val
		rewrittenAttr = append(rewrittenAttr, href)

		token.Attr = rewrittenAttr
	}
}

func rewriteScriptSrc(token *html.Token) {
	rewrittenAttr := []html.Attribute{}

	for _, attr := range token.Attr {
		switch {
		case attr.Key == "src":
			val := attr.Val
			if strings.HasPrefix(val, "//") {
				val = "https:" + val
			}
			val = os.Getenv("HEROKU_URL") + "/" + val
			attr.Val = val
			rewrittenAttr = append(rewrittenAttr, attr)
		default:
			rewrittenAttr = append(rewrittenAttr, attr)
		}
	}

	token.Attr = rewrittenAttr
}
