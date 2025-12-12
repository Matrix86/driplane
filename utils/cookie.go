package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

// Cookies maps the json file exported from Chrome containing the cookies
type Cookies []struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate,omitempty"`
	HostOnly       bool    `json:"hostOnly"`
	HTTPOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreID        string  `json:"storeId"`
	Value          string  `json:"value"`
	ID             int     `json:"id"`
}

// ParseCookieFile transforms JSON file in slice of http.Cookie
func ParseCookieFile(filename string) ([]*http.Cookie, error) {
	var cookies Cookies
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &cookies)
	if err != nil {
		return nil, err
	}

	httpCookies := make([]*http.Cookie, len(cookies))
	for i, c := range cookies {
		secs := int64(c.ExpirationDate)
		nsecs := int64((c.ExpirationDate - float64(secs)) * 1e9)

		httpCookies[i] = &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  time.Unix(secs, nsecs),
			Secure:   c.Secure,
			HttpOnly: c.HTTPOnly,
		}
	}

	return httpCookies, nil
}
