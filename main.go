// Copyright (c) 2015 Niklas Wolber
// This file is licensed under the MIT license.
// See the LICENSE file for more information.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nwolber/proxy/rrproxy"
)

const (
	endpointDefault = ":80"
	pathDefault     = "/"
)

func main() {
	var endpoint, path, cert, key string
	var help bool

	flag.StringVar(&endpoint, "ep", endpointDefault, "frontend endpoint")
	flag.StringVar(&path, "path", pathDefault, "frontend path")
	flag.StringVar(&cert, "cert", "", "PEM-encoded certificate file for HTTPS connections")
	flag.StringVar(&key, "key", "", "PEM-encoded key file for HTTPS connections")
	flag.BoolVar(&help, "h", false, "display help")
	flag.Parse()
	urls := flag.Args()

	if len(urls) == 0 || endpoint == "" || path == "" || help {
		printDefaults()
		return
	}

	if cert != "" && key == "" || cert == "" && key != "" {
		fmt.Println("certificate and key have to be provided")
		return
	}

	parsedURLs := make([]*url.URL, len(urls))
	for i, u := range urls {
		url, err := url.Parse(u)

		if err != nil {
			fmt.Println(u, "is not a valid URL")
			return
		}

		parsedURLs[i] = url
	}

	proxy, err := rrproxy.New(parsedURLs...)

	if err != nil {
		fmt.Println("unable to create proxy")
		fmt.Println(err)
		return
	}

	http.Handle(path, http.StripPrefix(path, proxy))

	if cert != "" {
		if endpoint == endpointDefault {
			endpoint = ":443"
		}
		err = http.ListenAndServeTLS(endpoint, cert, key, nil)
	} else {
		err = http.ListenAndServe(endpoint, nil)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func printDefaults() {
	fmt.Println("proxy [options] backend1 [backend2 ... [backendN]]")
	flag.PrintDefaults()
}
