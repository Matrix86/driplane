---
title: "WebRSS"
date: 2024-01-01T00:00:00+00:00
draft: false
---

## WebRSS

This feeder creates a stream by scraping a website and extracting articles using CSS selectors.
It is useful for websites that do not expose an RSS/Atom feed. It is based on [Colly](https://github.com/gocolly/colly) so you can refer to it for more info on how selectors work.

### Parameters

| Parameter           | Type                                                     | Default | Description                                                                 |
|---------------------|----------------------------------------------------------|---------|-----------------------------------------------------------------------------|
| **url**             | _STRING_                                                 | empty   | URL of the page to scrape                                                   |
| **item_selector**   | _STRING_                                                 | empty   | CSS selector matching each article/item block on the page                   |
| **title_selector**  | _STRING_                                                 | empty   | CSS selector for the article title (relative to the item block)             |
| **link_selector**   | _STRING_                                                 | empty   | CSS selector for the article link (relative to the item block)              |
| **desc_selector**   | _STRING_                                                 | empty   | CSS selector for the article description (relative to the item block)       |
| **date_selector**   | _STRING_                                                 | empty   | CSS selector for the article date (relative to the item block)              |
| **link_attr**       | _STRING_                                                 | "href"  | HTML attribute to read the URL from on the matched link element             |
| **freq**            | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | 60m     | how often the page should be scraped                                        |

{{< notice info "Example" >}}
`... | <webrss: url="https://website.io/blog", item_selector="a[href^='/blog/']", title_selector="h4, h3", link_selector="self", desc_selector="p", freq="1h"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the `title` of the article extracted via `title_selector`.

#### Extra

| Name         | Description                                                                 |
|--------------|-----------------------------------------------------------------------------|
| title        | title of the article extracted via `title_selector`                         |
| link         | absolute URL of the article extracted via `link_selector`                   |
| description  | description or excerpt of the article extracted via `desc_selector`         |
| published_at | publication date of the article extracted via `date_selector`               |

{{< notice warning "ATTENTION" >}}
`description` and `published_at` will be empty if the respective selectors (`desc_selector`, `date_selector`) are not configured or the element is not found on the page.
{{< /notice >}}

### Notes

- Relative URLs are automatically resolved to absolute URLs based on the scraped page URL.
- Already seen links are tracked in memory and will not be propagated again within the same run. This prevents duplicate messages across polling cycles.
- The selectors for `title_selector`, `link_selector`, `desc_selector` and `date_selector` are evaluated **relative to the element matched by `item_selector`**.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}}