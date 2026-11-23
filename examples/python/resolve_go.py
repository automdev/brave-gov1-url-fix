#!/usr/bin/env python3
"""Resolve a Brave Search /go/v1 URL by reading the Location header.

Web results wrap destinations as::

    https://search.brave.com/go/v1/<token>

The token is opaque. Do not decode it. Brave reveals the real URL in the
HTTP ``Location`` header, and only when ``Referer`` is the SERP the link
came from. ``https://search.brave.com/`` or a bare ``/search`` returns the
cloak page (200) and no ``Location``.

Do not follow the redirect (``allow_redirects=False``). Try HEAD first,
then GET.

Usage::

    python3 examples/python/resolve_go.py \
        'https://search.brave.com/go/v1/TOKEN' \
        'https://search.brave.com/search?q=piloterr+api&source=web'
"""

from __future__ import annotations

import sys
import urllib.parse

import requests

ORIGIN = "https://search.brave.com"


def is_brave_go(url: str) -> bool:
    """Return True if *url* is an https://search.brave.com/go/v1/… link."""
    try:
        parsed = urllib.parse.urlparse(url)
    except ValueError:
        return False
    return (
        parsed.scheme == "https"
        and parsed.netloc == "search.brave.com"
        and parsed.path.startswith("/go/v1/")
        and len(parsed.path) > len("/go/v1/")
    )


def resolve_go(url: str, referer: str, timeout: int = 10) -> str | None:
    """Return the destination from Location, or None if Brave sent none."""
    if url.startswith("/go/v1/"):
        url = ORIGIN + url
    if not is_brave_go(url):
        raise ValueError("expected https://search.brave.com/go/v1/<token>")
    if not referer.startswith("https://search.brave.com/search?"):
        raise ValueError("Referer must be the SERP URL, with its query string")

    headers = {"Referer": referer}
    for method in ("head", "get"):
        response = requests.request(
            method,
            url,
            allow_redirects=False,
            headers=headers,
            timeout=timeout,
        )
        location = response.headers.get("Location")
        if not location:
            continue
        dest = urllib.parse.urljoin(url, location)
        # A cloak page can redirect back onto /go/v1. That is not a destination.
        if is_brave_go(dest):
            continue
        return dest
    return None


def main() -> int:
    if len(sys.argv) != 3:
        print(
            "Usage: resolve_go.py 'https://search.brave.com/go/v1/TOKEN' "
            "'https://search.brave.com/search?q=...'",
            file=sys.stderr,
        )
        return 2
    dest = resolve_go(sys.argv[1], sys.argv[2])
    if not dest:
        print("no Location header", file=sys.stderr)
        return 1
    print(dest)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
