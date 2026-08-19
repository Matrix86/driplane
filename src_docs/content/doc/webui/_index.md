---
title: "Web interface"
weight: 6
---

driplane ships an optional web interface, served by the daemon itself, to
create and edit rules, templates and JavaScript plugins, validate them, and
watch the running pipelines.

## Enabling it

```yaml
web:
  enable: true
  address: "127.0.0.1:8080"
  token: ""
```

Or from the command line:

```bash
driplane -config config.yaml -web
driplane -config config.yaml -web-address 127.0.0.1:9000
```

`-web-address` enables the interface on its own — you do not need to also
pass `-web`.

When `token` is empty a random one is generated at every startup and printed in
the log together with the URL to open:

```
[2026-08-19 10:00:00] imp web interface on http://127.0.0.1:8080/?token=Xy...
```

## Security

The interface writes the files that the daemon executes: rules, templates and
JS plugins. Whoever reaches it can run code on the host.

- It binds to `127.0.0.1` by default. Binding it anywhere else logs a warning.
- Every `/api/*` endpoint requires the token, with no exceptions: there is no
  anonymous access to any data or action. The static UI shell — the HTML,
  CSS, JS and vendored CodeMirror files — is served on `GET` without a token,
  because a browser cannot attach one to `<link>` and `<script src>` requests
  and the UI could not load otherwise. Nothing under that exemption is
  per-installation data: it is only the page itself and its vendored editor
  library, the same bytes for every install.
- To expose it outside the machine, put it behind a reverse proxy handling TLS
  and, if needed, an additional authentication layer.
- The web package's own code never reads or writes `config.yaml`: it does not
  parse feeder credentials out of it or serve them back. It does, however,
  hand the daemon's real config file path to the `/api/test` subprocess, so
  that a rule test runs against the same configuration the live daemon uses
  and behaves like the real thing — including whatever `general.debug`
  setting is in effect on that file.

## What it does

- **Editor** — rules, templates and JS plugins, with syntax highlighting for the
  rule DSL, autocompletion of filter and feeder names, and syntax validation
  while typing. Saving keeps the previous version in a `.bak` file next to it.
- **Test** — runs the rule currently open through a dry-run of the driplane
  binary in a temporary copy of the rules directory, so the running daemon is
  never touched.
- **Runtime** — the loaded rules with their message counters, the daemon state,
  the live log, plus the Reload, Start and Stop buttons. These three requests
  only ask the daemon to act; they answer as soon as the request is accepted,
  before the rebuild has actually happened, so a `200` response means
  "requested", not "succeeded". Watch the state indicator (it polls every 3
  seconds) to see whether a reload actually went through. If it did not — a
  rule with a syntax error, for instance — the daemon does **not** keep
  running the previous rules: it stops collecting entirely and stays stopped
  until a valid reload arrives, because feeders are always stopped before the
  rebuild is attempted.
