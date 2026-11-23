# search.brave.com/go/v1: read `Location`

Brave Search no longer puts the destination in the organic `href`. Web
results go through `https://search.brave.com/go/v1/<token>`. That token is
**not** an encoded URL: you cannot decode it. Brave only reveals the
destination in the HTTP **`Location`** header, and only when `Referer` is
the SERP the link came from.

This repo explains the method, with **curl**, **Python**, **Node**, and **Go**
examples: `GET` or `HEAD` without following the redirect, then read
`Location`.

Autom article: [How to resolve search.brave.com/go/v1](https://www.autom.dev/blog/brave-gov1-url-fix)

## The idea in one sentence

**Do not follow the redirect. Send the SERP as `Referer`. Read `Location`.**

```
GET /go/v1/<token> HTTP/1.1
Host: search.brave.com
Referer: https://search.brave.com/search?q=piloterr+api&source=web

HTTP/1.1 302 Found
Location: https://www.piloterr.com/library
```

`HEAD` asks for headers only (no body). `GET` with `redirect: "manual"` (or
`allow_redirects=False`) does the same read. Both work when the `Referer` is
the full SERP URL, query string included.

`Referer: https://search.brave.com/` or a bare
`https://search.brave.com/search` returns the cloak page (HTTP 200) and no
`Location`.

Images and videos still use direct `https` links. This change is the web SERP.

## curl: HEAD, do not follow

```bash
curl -sI --max-redirs 0 \
  -H 'Referer: https://search.brave.com/search?q=piloterr+api&source=web' \
  'https://search.brave.com/go/v1/TOKEN' \
  | grep -i '^location:'
```

`-I` = HEAD. Without `-L`, curl does not follow to the final page.
`--max-redirs 0` blocks any follow. The `Location:` line is the real URL.

Same thing with GET:

```bash
curl -sD - -o /dev/null --max-redirs 0 \
  -H 'Referer: https://search.brave.com/search?q=piloterr+api&source=web' \
  'https://search.brave.com/go/v1/TOKEN' \
  | grep -i '^location:'
```

Ready-to-run script: [`examples/curl/head.sh`](examples/curl/head.sh).

## Python

```python
import requests

url = "https://search.brave.com/go/v1/TOKEN"
serp = "https://search.brave.com/search?q=piloterr+api&source=web"
r = requests.head(
    url,
    allow_redirects=False,
    headers={"Referer": serp},
    timeout=10,
)
print(r.status_code, r.headers.get("Location"))
```

If `HEAD` returns nothing, retry with `GET` and do not follow:

```python
r = requests.get(url, allow_redirects=False, headers={"Referer": serp}, timeout=10)
print(r.headers.get("Location"))
```

File: [`examples/python/resolve_go.py`](examples/python/resolve_go.py).

## Node

Use `fetch(..., { redirect: "manual" })` then
`response.headers.get("location")`. In Node, the headers are visible.

```javascript
const url = "https://search.brave.com/go/v1/TOKEN";
const serp = "https://search.brave.com/search?q=piloterr+api&source=web";
const res = await fetch(url, {
  method: "HEAD",
  redirect: "manual",
  headers: { Referer: serp },
});
console.log(res.status, res.headers.get("location"));
```

File: [`examples/javascript/resolve_go.js`](examples/javascript/resolve_go.js).

## Go

`http.Client` follows redirects by default. Return
`http.ErrUseLastResponse` from `CheckRedirect` so the first hop stays
in place and `Location` can be read.

```go
client := &http.Client{
    CheckRedirect: func(*http.Request, []*http.Request) error {
        return http.ErrUseLastResponse
    },
}
req, _ := http.NewRequest(http.MethodHead, goURL, nil)
req.Header.Set("Referer", serp)
res, _ := client.Do(req)
fmt.Println(res.StatusCode, res.Header.Get("Location"))
```

File: [`examples/go/resolve_go.go`](examples/go/resolve_go.go).

## Why not decode the token?

The `/go/v1/<token>` value is not a URL and not a public encoding of one.
Brave resolves it when the request looks like a click from the SERP, and
puts the destination in `Location`. Do not decode the token: request
`/go/v1/<token>` with that `Referer` and read the redirect.

## At Autom

Autom’s [Brave Search](https://www.autom.dev/library/brave-search) API
resolves these `/go/v1` URLs on the server (same idea: no follow, SERP as
`Referer`, read `Location`) and already returns the final URL in
`organic_results[].link`.

```bash
curl -s https://api.autom.dev/v1/brave/search \
  -H "x-api-key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query":"piloterr api","gl":"us","hl":"en"}'
```

1,000 free requests: [autom.dev/pricing](https://www.autom.dev/pricing).

## License

MIT.
