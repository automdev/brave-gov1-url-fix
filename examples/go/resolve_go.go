// Resolve a Brave Search /go/v1 URL by reading the Location header.
//
// Web results wrap destinations as:
//
//	https://search.brave.com/go/v1/<token>
//
// The token is opaque. Do not decode it. Brave reveals the real URL in the
// HTTP Location header, and only when Referer is the SERP the link came
// from. https://search.brave.com/ or a bare /search returns the cloak page
// (200) and no Location.
//
// Do not follow the redirect (CheckRedirect returns ErrUseLastResponse).
// Try HEAD first, then GET.
//
// Usage:
//
//	go run ./examples/go 'https://search.brave.com/go/v1/TOKEN' 'https://search.brave.com/search?q=...'
package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const origin = "https://search.brave.com"

func isBraveGo(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme == "https" &&
		u.Hostname() == "search.brave.com" &&
		strings.HasPrefix(u.Path, "/go/v1/") &&
		len(u.Path) > len("/go/v1/")
}

func resolveGo(raw, referer string) (string, error) {
	if strings.HasPrefix(raw, "/go/v1/") {
		raw = origin + raw
	}
	if !isBraveGo(raw) {
		return "", errors.New("expected https://search.brave.com/go/v1/<token>")
	}
	if !strings.HasPrefix(referer, "https://search.brave.com/search?") {
		return "", errors.New("Referer must be the SERP URL, with its query string")
	}

	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, method := range []string{http.MethodHead, http.MethodGet} {
		req, err := http.NewRequest(method, raw, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Referer", referer)

		res, err := client.Do(req)
		if err != nil {
			return "", err
		}
		location := res.Header.Get("Location")
		_ = res.Body.Close()
		if location == "" {
			continue
		}
		abs, err := res.Request.URL.Parse(location)
		if err != nil {
			return "", err
		}
		if isBraveGo(abs.String()) {
			continue
		}
		return abs.String(), nil
	}
	return "", errors.New("no Location header")
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run . 'https://search.brave.com/go/v1/TOKEN' 'https://search.brave.com/search?q=...'")
		os.Exit(2)
	}
	dest, err := resolveGo(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(dest)
}
