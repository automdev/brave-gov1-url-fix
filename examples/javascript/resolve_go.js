#!/usr/bin/env node
/**
 * Resolve a Brave Search /go/v1 URL by reading the Location header.
 *
 * Web results wrap destinations as:
 *   https://search.brave.com/go/v1/<token>
 * The token is opaque. Do not decode it. Brave reveals the real URL in the
 * HTTP Location header, and only when Referer is the SERP the link came
 * from. https://search.brave.com/ or a bare /search returns the cloak page
 * (200) and no Location.
 *
 * Use fetch with redirect: "manual". Do not follow the hop. Try HEAD first,
 * then GET. In Node, Location is visible.
 *
 * Usage:
 *   node examples/javascript/resolve_go.js \
 *     'https://search.brave.com/go/v1/TOKEN' \
 *     'https://search.brave.com/search?q=piloterr+api&source=web'
 */

const ORIGIN = "https://search.brave.com";

function isBraveGo(raw) {
  try {
    const url = new URL(raw);
    return (
      url.protocol === "https:" &&
      url.hostname === "search.brave.com" &&
      url.pathname.startsWith("/go/v1/") &&
      url.pathname.length > "/go/v1/".length
    );
  } catch {
    return false;
  }
}

async function resolveGo(raw, referer) {
  const url = raw.startsWith("/go/v1/") ? ORIGIN + raw : raw;
  if (!isBraveGo(url)) {
    throw new Error("expected https://search.brave.com/go/v1/<token>");
  }
  if (!referer.startsWith("https://search.brave.com/search?")) {
    throw new Error("Referer must be the SERP URL, with its query string");
  }

  const headers = { Referer: referer };
  for (const method of ["HEAD", "GET"]) {
    const res = await fetch(url, { method, redirect: "manual", headers });
    const location = res.headers.get("location");
    if (!location) continue;
    const dest = new URL(location, url).href;
    if (isBraveGo(dest)) continue;
    return dest;
  }
  return null;
}

const input = process.argv[2];
const referer = process.argv[3];
if (!input || !referer) {
  console.error(
    "Usage: node resolve_go.js 'https://search.brave.com/go/v1/TOKEN' 'https://search.brave.com/search?q=...'",
  );
  process.exit(2);
}

resolveGo(input, referer)
  .then((dest) => {
    if (!dest) {
      console.error("no Location header");
      process.exit(1);
    }
    console.log(dest);
  })
  .catch((err) => {
    console.error(err.message);
    process.exit(1);
  });
