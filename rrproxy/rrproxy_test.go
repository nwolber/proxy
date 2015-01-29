// Copyright (c) 2015 Niklas Wolber
// This file is licensed under the MIT license.
// See the LICENSE file for more information.
package rrproxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSadPath(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Error("Expected an error, but didn't get any.")
	}
}

type proxy struct {
	c int
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.c++
}

var tests = []struct {
	numProxies  int
	numRequests int
	result      []int
}{
	{numProxies: 1, numRequests: 4, result: []int{4}},
	{numProxies: 2, numRequests: 4, result: []int{2, 2}},
	{numProxies: 3, numRequests: 10, result: []int{4, 3, 3}},
}

func TestRoundRobinReverseProxy(t *testing.T) {
	for _, tt := range tests {
		func() {
			proxies := make([]*proxy, tt.numProxies)
			urls := make([]*url.URL, tt.numProxies)

			for i := 0; i < tt.numProxies; i++ {
				p := &proxy{}
				proxies[i] = p
				server := httptest.NewServer(p)
				defer server.Close()
				url, _ := url.Parse(server.URL)
				urls[i] = url
			}

			p, err := New(urls...)
			defer p.Close()

			if err != nil {
				t.Errorf("expected NewRoundRobinReverseProxy not to fail, but did so: %s", err)
			}

			frontend := httptest.NewServer(p)

			for i := 0; i < tt.numRequests; i++ {
				req, err := http.NewRequest("GET", frontend.URL, nil)

				if err != nil {
					t.Errorf("expected request not to fail, but did so: %s", err)
				}

				http.DefaultClient.Do(req)
			}

			for i := 0; i < tt.numProxies; i++ {
				if tt.result[i] != proxies[i].c {
					t.Errorf("expected proxy %d to receive %d requests, but got %d", i, tt.result[i], proxies[i].c)
				}
			}
		}()
	}
}

var rewriteURLTests = []struct {
	frontend string
	backend  string
	want     string
}{
	// scheme
	{frontend: "http://a/", backend: "http://b/", want: "http://b/"},
	{frontend: "http://a/", backend: "https://b/", want: "https://b/"},
	//path
	{frontend: "http://a/path", backend: "http://b/", want: "http://b/"},
	{frontend: "http://a/", backend: "http://b/path", want: "http://b/path"},
	// query
	{frontend: "http://a?param=value", backend: "http://b", want: "http://b?param=value"},
	{frontend: "http://a/", backend: "http://b?param=value", want: "http://b?param=value"},
	{frontend: "http://a?paramA=value1", backend: "http://b?paramB=value2", want: "http://b?paramA=value1&paramB=value2"},
	// username/passowrd
	{frontend: "http://user:pass@a/", backend: "http://b/", want: "http://user:pass@b/"},
	{frontend: "http://a/", backend: "http://user:pass@b/", want: "http://user:pass@b/"},
	{frontend: "http://userA:passA@a/", backend: "http://userB:passB@b/", want: "http://userA:passA@b/"},
	// username
	{frontend: "http://user@a/", backend: "http://b/", want: "http://user@b/"},
	{frontend: "http://a/", backend: "http://user@b/", want: "http://user@b/"},
	{frontend: "http://userA@a/", backend: "http://userBB@b/", want: "http://userA@b/"},
}

func TestRewriteURL(t *testing.T) {
	for _, tt := range rewriteURLTests {
		request, _ := url.Parse(tt.frontend)
		target, _ := url.Parse(tt.backend)

		rewriteURL(request, target)

		if request.String() != tt.want {
			t.Errorf("want: %s, got: %s", tt.want, request)
		}
	}
}
