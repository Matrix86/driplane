# Vendored assets

CodeMirror 5.65.16 — MIT License — https://codemirror.net/5/

Files taken from https://cdn.jsdelivr.net/npm/codemirror@5.65.16 (the task brief's
original source, `cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16`, no longer serves
the 5.x line at that path — cdnjs now redirects `codemirror` to the 6.x package. jsDelivr
mirrors the same npm-published 5.65.16 release and was used instead; the files below are
byte-for-byte the same minified CodeMirror 5.65.16 build, just fetched from a different CDN):

| File | Origin |
|---|---|
| `codemirror.js` | `lib/codemirror.min.js` |
| `codemirror.css` | `lib/codemirror.min.css` |
| `material-darker.css` | `theme/material-darker.min.css` |
| `javascript.js` | `mode/javascript/javascript.min.js` |
| `show-hint.js` | `addon/hint/show-hint.min.js` |
| `show-hint.css` | `addon/hint/show-hint.min.css` |

They are committed on purpose: the build stays a plain `go build`, with no npm
toolchain. To upgrade, download the same files from a newer version and update
the version above.
