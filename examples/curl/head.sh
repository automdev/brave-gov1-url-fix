#!/usr/bin/env bash
# Resolve a Brave Search /go/v1 URL by reading the Location header.
#
# Web results wrap destinations as:
#   https://search.brave.com/go/v1/<token>
# The token is opaque. Do not decode it. Brave reveals the real URL in the
# HTTP Location header, and only when Referer is the SERP the link came from.
# A Referer of https://search.brave.com/ or a bare /search returns the cloak
# page (200) and no Location. Do not follow the redirect (-L).
#
# Usage:
#   ./examples/curl/head.sh \
#     'https://search.brave.com/go/v1/TOKEN' \
#     'https://search.brave.com/search?q=piloterr+api&source=web'

set -euo pipefail

GO="${1:-}"
REFERER="${2:-}"
if [[ -z "$GO" || "$GO" != https://search.brave.com/go/v1/* || -z "$REFERER" ]]; then
  echo "Usage: $0 'https://search.brave.com/go/v1/TOKEN' 'https://search.brave.com/search?q=...'" >&2
  exit 1
fi

# HEAD: headers only, no body. --max-redirs 0 refuses any follow.
echo "== HEAD =="
curl -sI --max-redirs 0 \
  -H "Referer: $REFERER" \
  "$GO" | grep -iE '^(HTTP/|location:)'

echo
# GET without -L: same Location read, no follow.
echo "== GET (no follow) =="
curl -sD - -o /dev/null --max-redirs 0 \
  -H "Referer: $REFERER" \
  "$GO" | grep -iE '^(HTTP/|location:)'
