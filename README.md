<p align="center">
  <img src="https://github.com/Matrix86/driplane/raw/gh-pages/logo.png" alt="Driplane Logo" width="300"/>
</p>

<p align="center">
  <a href="https://github.com/Matrix86/driplane/blob/master/LICENSE.md"><img src="https://img.shields.io/github/license/Matrix86/driplane" alt="License"/></a>
  <a href="https://github.com/Matrix86/driplane"><img src="https://img.shields.io/github/go-mod/go-version/Matrix86/driplane" alt="Go version"/></a>
  <a href="https://github.com/Matrix86/driplane/actions"><img src="https://img.shields.io/github/workflow/status/Matrix86/driplane/Build%20and%20Test/master" alt="Build Status"/></a>
  <a href="https://github.com/Matrix86/driplane/releases/latest"><img src="https://img.shields.io/github/v/release/Matrix86/driplane?color=red" alt="Latest Release"/></a>
  <a href="https://codecov.io/gh/Matrix86/driplane"><img src="https://img.shields.io/codecov/c/github/Matrix86/driplane" alt="Coverage"/></a>
  <a href="https://www.buymeacoffee.com/mtx86"><img src="https://img.shields.io/badge/Buy%20me%20a-%F0%9F%8D%BA%20beer-blue?style=flat-square&color=grey&labelColor=blue&logo=buymeacoffee&logoColor=white" alt="Buy me a beer"/></a>
</p>

<h3 align="center">Event-driven automation pipelines — monitor anything, react to everything.</h3>

---

**Driplane** is a lightweight, rule-based automation engine written in Go. You define *what* to watch (RSS feeds, Twitter/X, Slack, IMAP, files, websites…) and *what to do* when something interesting happens (send a Telegram/Slack alert, write to a file, call a webhook, run custom JS logic, and more). Everything is wired together using a simple pipeline syntax inspired by Unix pipes.

```
RSS feed ──► filter by keyword ──► deduplicate ──► format message ──► send to Telegram
```

---

## ✨ Key Features

- **Declarative pipeline rules** — a clean, pipe-based mini-language to wire sources to actions
- **Many built-in feeders** — RSS, Twitter/X, Slack, Telegram, IMAP, file, folder, web scraping, APT, timer
- **Rich filter library** — text matching, regex, caching, deduplication, hashing, JSON/HTML parsing, HTTP calls, rate limiting, LLM integration, and more
- **JavaScript plugin system** — extend driplane with custom logic using embedded JS
- **Composable rules** — reuse and chain rules with `@rule_name` references
- **Template support** — format output with Go templates and external template files
- **Docker-ready** — run it as a container with minimal configuration

---

## 🔧 How It Works

A **rule** is the basic unit of driplane. It starts with a **feeder** that produces a stream of messages, and passes each message through a chain of **filters**.

```
RuleName => <feeder_type: param="value"> | filter1(...) | filter2(...) | filter3(...) ;
```

Each **filter** decides whether to pass the message along or drop it. Filters can also transform the message — extracting fields, reformatting text, or calling external services.

Rules can be reused inside other rules using the `@` prefix:

```
# Define a reusable action
notify => http(url="https://hooks.slack.com/...", method="POST", rawData="{{.main}}");

# Use it inside another rule
my_pipeline => <rss: url="https://example.com/feed.rss", freq="5m"> |
               text(pattern="security", regexp="true") |
               cache(ttl="24h") |
               format(template="New article: {{ .link }}") |
               @notify;
```

### The Message Object

Every piece of data flowing through a pipeline is a **Message**. It has:
- A **main text** field (referenced as `{{ .main }}` in templates)
- **Extra fields** set by feeders and filters (e.g. `.link`, `.title`, `.channel`, `.rule_name`, `.source_feeder`)

Filters operate on the main text by default, or on any extra field via the `target` parameter.

---

## 📦 Installation

### Download a release

Grab the latest binary from the [Releases page](https://github.com/Matrix86/driplane/releases).

### Build from source

```bash
git clone https://github.com/Matrix86/driplane.git
cd driplane
make build
```

### Docker

```bash
docker pull ghcr.io/matrix86/driplane:latest
docker run -v /path/to/config.yml:/config.yml \
           -v /path/to/rules:/rules \
           ghcr.io/matrix86/driplane:latest -config /config.yml -rules /rules
```

---

## 🚀 Quick Start

**1. Create a config file** (`config.yml`):

```yaml
# Credentials and global settings go here
# See: https://matrix86.github.io/driplane/doc/configuration/
```

**2. Write a rule file** (`rules/my_rule.rule`):

```
# Monitor a RSS feed and send new articles to Telegram
Feed => <rss: url="https://feeds.example.com/news.rss", freq="5m">;

news_alert => @Feed |
              cache(ttl="48h", target="link") |
              format(template="📰 {{ .title }}\n{{ .link }}") |
              http(url="https://api.telegram.org/botTOKEN/sendMessage",
                   method="POST",
                   rawData="{\"chat_id\": \"CHAT_ID\", \"text\": \"{{.main}}\"}");
```

**3. Run driplane**:

```bash
driplane -config config.yml -rules ./rules
```

---

## 📋 Rule Syntax Reference

| Concept | Syntax | Description |
|---|---|---|
| Rule definition | `Name => ... ;` | Define a named rule |
| Feeder | `<type: key="value">` | Source of data (must be first in a rule) |
| Filter | `filter(key="value")` | Process and/or filter messages |
| NOT modifier | `!filter(...)` | Negate a filter — drop the message if condition is met |
| Rule call | `@RuleName` | Inline another rule as a filter step |
| Import | `#import "file.rule"` | Include rules from another file |
| Template | `{{ .main }}`, `{{ .field }}` | Reference message fields in strings |

---

## 📡 Feeders

Feeders are the **data sources**. Each feeder runs independently and feeds messages into its pipeline.

| Feeder | Description |
|---|---|
| `rss` | Poll an RSS/Atom feed at a defined interval |
| `webrss` | Monitor a web page and expose it as a feed |
| `web` | Fetch and monitor a web page |
| `twitter` | Stream tweets by keyword or user |
| `slack` | Listen to Slack events (messages, file uploads…) |
| `telegram` | Receive Telegram bot messages |
| `imap` | Monitor an IMAP mailbox |
| `file` | Watch a file for changes |
| `folder` | Watch a folder for new/changed files |
| `apt` | Monitor APT package updates |
| `timer` | Trigger a pipeline on a schedule |

---

## 🔀 Filters

Filters are the **processing units**. They run in sequence and each one decides whether to pass the message forward or drop it.

| Category | Filters |
|---|---|
| **Text & matching** | `text`, `regex`, `hash`, `striptag`, `url` |
| **Data parsing** | `json`, `html`, `pdf`, `mimetype`, `xls` |
| **Flow control** | `cache`, `changed`, `ratelimit`, `random` |
| **Transformation** | `format`, `override`, `number` |
| **Actions** | `http`, `mail`, `file`, `echo`, `system` |
| **Integrations** | `slack`, `telegram`, `elasticsearch`, `llm` |
| **Custom logic** | `js` (JavaScript plugin) |

### Negating a filter

Prefix any filter with `!` to invert its behavior — the message is **dropped** if the condition is met:

```
# Drop messages in Spanish, keep everything else
!text(target="language", pattern="es")
```

---

## 💡 Examples

### Monitor RSS for keywords → Telegram alert

```
Feed => <rss: url="https://feeds.example.com/security.rss", freq="10m">;

security_news => @Feed |
                 cache(ttl="48h", target="link") |
                 text(pattern="(?i)vuln|exploit|CVE", regexp="true", target="title") |
                 format(template="🚨 *{{ .title }}*\n{{ .link }}") |
                 http(url="https://api.telegram.org/botTOKEN/sendMessage",
                      method="POST",
                      headers="{\"Content-type\": \"application/json\"}",
                      rawData="{\"chat_id\":\"CHATID\",\"text\":\"{{.main}}\",\"parse_mode\":\"Markdown\"}");
```

---

### Twitter/X → Slack webhook (with deduplication)

Watch for tweets mentioning malware keywords, extract hashes, deduplicate with a 24h cache, and post to Slack:

```
Twitter => <twitter: users="user1, user2", keywords="malware, ransomware">;

slack_notify => http(url="https://hooks.slack.com/services/XXX/YYY/ZZZ",
                     method="POST",
                     headers="{\"Content-type\": \"application/json\"}",
                     rawData="{{.main}}");

tweet_pipeline => @Twitter |
                  !text(target="language", pattern="es") |
                  hash(extract="true") |
                  override(name="hash", value="{{ .main }}") |
                  cache(ttl="24h", global="true") |
                  format(file="slack_template.txt") |
                  @slack_notify;
```

---

### Simple Slack bot — analyze uploaded files

A bot that responds to file uploads in Slack by running custom JS analysis logic:

```
SlackEvent => <slack>;

file_analysis => @SlackEvent |
                 text(target="type", pattern="file_share") |
                 slack(action="download_file", filename="/tmp/{{ .name }}") |
                 js(path="analyzer.js", function="AnalyzeFile") |
                 format(file="report_template.txt") |
                 slack(action="send_message", to="{{.channel}}", target="main");
```

---

### File monitoring → email alert

Watch a sensitive file and send an email if it changes:

```
ConfigFile => <file: path="/etc/important.conf">;

config_changed => @ConfigFile |
                  changed() |
                  format(template="⚠️ File changed at {{ .timestamp }}") |
                  mail(to="admin@example.com", subject="Config file modified!", body="{{.main}}");
```

---

## ⚙️ CLI Options

```
Usage: driplane [options]

  -config string    Path to the configuration file
  -rules  string    Path to the rules directory
  -js     string    Path to the JS plugins directory
  -debug            Enable verbose debug logging
  -dry-run          Parse and validate rules without running them
  -help             Show this help message
```

---

## 📚 Documentation

Full documentation is available at **[https://matrix86.github.io/driplane/doc/](https://matrix86.github.io/driplane/doc/)**, including:

- [Installation & Docker setup](https://matrix86.github.io/driplane/doc/installation/)
- [Configuration reference](https://matrix86.github.io/driplane/doc/configuration/)
- [Rule syntax guide](https://matrix86.github.io/driplane/doc/rules/syntax/)
- [All feeders with parameters](https://matrix86.github.io/driplane/doc/feeders/)
- [All filters with parameters](https://matrix86.github.io/driplane/doc/filters/)
- [JavaScript plugin system](https://matrix86.github.io/driplane/doc/filters/js/basics/)

---

## 🤝 Contributing

Contributions are welcome! Feel free to open issues, suggest new feeders or filters, or submit pull requests.

---

## 📄 License

Driplane is released under the [GPL-3.0 License](LICENSE.md).

---

<p align="center">
  If you find driplane useful, consider <a href="https://www.buymeacoffee.com/mtx86">buying me a beer 🍺</a>
</p>