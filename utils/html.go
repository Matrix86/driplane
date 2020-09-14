package utils

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

type HTMLMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
}

func GetMetaFromHtml(s string) *HTMLMeta {
	z := html.NewTokenizer(strings.NewReader(s))
	titleFound := false
	hm := &HTMLMeta{}

	for {
		tt := z.Next()
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:

			t := z.Token()
			if t.Data == `body` {
				return hm
			}
			if t.Data == "title" {
				titleFound = true
			}
			if t.Data == "meta" {
				desc, ok := extractMetaProperty(t, "description")
				if ok {
					hm.Description = desc
				}

				ogTitle, ok := extractMetaProperty(t, "og:title")
				if ok {
					hm.Title = ogTitle
				}

				ogDesc, ok := extractMetaProperty(t, "og:description")
				if ok {
					hm.Description = ogDesc
				}

				ogImage, ok := extractMetaProperty(t, "og:image")
				if ok {
					hm.Image = ogImage
				}

				ogSiteName, ok := extractMetaProperty(t, "og:site_name")
				if ok {
					hm.SiteName = ogSiteName
				}
			}
		case html.TextToken:
			if titleFound {
				t := z.Token()
				hm.Title = t.Data
				titleFound = false
			}
		case html.ErrorToken:
			return hm
		}
	}
}

func extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
	for _, attr := range t.Attr {
		if attr.Key == "property" && attr.Val == prop {
			ok = true
		}

		if attr.Key == "content" {
			content = attr.Val
		}
	}

	return
}

func ExtractTextFromHtml(s string) string {
	ret := ""
	domDocTest := html.NewTokenizer(strings.NewReader(s))
	previousStartTokenTest := domDocTest.Token()
	for {
		tt := domDocTest.Next()
		switch {
		case tt == html.ErrorToken:
			return ret // End of the document,  done
		case tt == html.StartTagToken:
			previousStartTokenTest = domDocTest.Token()
		case tt == html.TextToken:
			if previousStartTokenTest.Data == "script" ||
				previousStartTokenTest.Data == "noscript" ||
				previousStartTokenTest.Data == "style" {
				continue
			}
			TxtContent := strings.TrimSpace(html.UnescapeString(string(domDocTest.Text())))
			if len(TxtContent) > 0 {
				ret = fmt.Sprintf("%s %s", ret, TxtContent)
			}
		}
	}
}
