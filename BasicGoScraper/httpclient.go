package main

import (
	"net/http"
	"net/url"
)

func NewHttpClient() *http.Client {
	client := &http.Client{
		Jar: new(ClientJar),
	}
	return client
}

// Type ClientJar
type ClientJar struct {
	cookies map[string][]*http.Cookie
}

func (c *ClientJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if c.cookies == nil {
		c.cookies = make(map[string][]*http.Cookie)
	}
	c.cookies[u.Host] = cookies
}

func (c *ClientJar) Cookies(u *url.URL) []*http.Cookie {
	return c.cookies[u.Host]
}
