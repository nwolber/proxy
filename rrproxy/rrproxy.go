// Copyright (c) 2015 Niklas Wolber
// This file is licensed under the MIT license.
// See the LICENSE file for more information.

// Package rrproxy provides a reverse proxy which distributes requests round-robin on a number of backend servers.
package rrproxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// A RoundRobinReverseProxy distributes HTTP requests evenly on a number of hosts.
type RoundRobinReverseProxy struct {
	httputil.ReverseProxy
	close chan struct{}
}

// New creates an HTTP proxy, which distributes requests
// evenly on the given hosts.
func New(hosts ...*url.URL) (*RoundRobinReverseProxy, error) {
	closing := make(chan struct{})

	switch n := len(hosts); {
	case n <= 0:
		return nil, errors.New("need at least one host to proxy to")
	case n == 1:
		p := httputil.NewSingleHostReverseProxy(hosts[0])
		return &RoundRobinReverseProxy{ReverseProxy: *p, close: nil}, nil
	}

	c := make(chan *url.URL, len(hosts))
	director := func(req *http.Request) {
		target, ok := <-c

		if !ok {
			// proxy has been closed
			return
		}

		rewriteURL(req.URL, target)
	}

	go func() {
		for {
			for _, url := range hosts {
				select {
				case c <- url:
					continue
				case <-closing:
					close(c)
					close(closing)
					return
				}
			}
		}
	}()

	return &RoundRobinReverseProxy{
		ReverseProxy: httputil.ReverseProxy{Director: director},
		close:        closing,
	}, nil
}

// Close closes the proxy and deallocates resources.
func (p *RoundRobinReverseProxy) Close() {
	if p.close != nil {
		p.close <- struct{}{}
	}
}

func rewriteURL(req, target *url.URL) {
	req.Scheme = target.Scheme
	req.Host = target.Host
	req.Path = target.Path

	if target.RawQuery == "" || req.RawQuery == "" {
		req.RawQuery = req.RawQuery + target.RawQuery
	} else {
		req.RawQuery = req.RawQuery + "&" + target.RawQuery
	}

	if req.User == nil || req.User.Username() == "" {
		req.User = target.User
	}
}
