package utils

import (
	"reflect"
	"testing"
)

func TestGetMetaFromHTML(t *testing.T) {
	html := "<html><head><title>test</title><meta charset=\"UTF-8\"><meta name=\"description\" content=\"description\"/><meta name=\"keywords\" content=\"one, two, three\"><meta name=\"author\" content=\"John Doe\"><meta name=\"og:description\" content=\"og_description\"><meta name=\"og:title\" content=\"og_title\"><meta name=\"og:image\" content=\"og_image\"><meta name=\"og:site_name\" content=\"og_site_name\"></head></html>"
	type Test struct {
		Name     string
		Html     string
		Expected *HTMLMeta
	}
	tests := []Test{
		{"CorrectHtml", html, &HTMLMeta{Title: "og_title", Description: "og_description", Image: "og_image", SiteName: "og_site_name"}},
		{"WrongHtml", "<html><head><body></head></html>", &HTMLMeta{}},
		{"EmptyHtml", "", &HTMLMeta{}},
	}

	for _, v := range tests {
		had := GetMetaFromHTML(v.Html)
		if reflect.DeepEqual(had, v.Expected) == false {
			t.Errorf("%s: wrong parsing: expected=%#v had=%#v", v.Name, v.Expected, had)
		}
	}
}

func TestExtractTextFromHTML(t *testing.T) {
	html := "<html><head><title>test</title><meta charset=\"UTF-8\"></head><body><script type=\"text/javascript\">alert('test');</script><h1>title</h1> this is a text!</body></html>"
	type Test struct {
		Name     string
		Html     string
		Expected string
	}
	tests := []Test{
		{"CorrectHtml", html, " test title this is a text!"},
		{"NoTextHtml", "<html><head><body></head></html>", ""},
		{"EmptyHtml", "", ""},
	}

	for _, v := range tests {
		had := ExtractTextFromHTML(v.Html)
		if had != v.Expected {
			t.Errorf("%s: wrong parsing: expected=%#v had=%#v", v.Name, v.Expected, had)
		}
	}
}
