# gomost

[![Go Report Card](https://goreportcard.com/badge/github.com/landonia/gomost)](https://goreportcard.com/report/github.com/landonia/gomost)

A proxy allowing you to host multiple sites (unique hosts) from the one server

## Overview

gomost (GO Multiple hOST) is used when you need to host multiple sites each with different host names from one machine.
It is a simple reverse proxy for each unique host that can either deliver static content from the filesystem (it matches the host to a folder name), forward the request to another server endpoint or if integrating into your own app a custom handler for any host.

## Installation

With a healthy Go Language installed, simply run `go get github.com/landonia/gomost`

### Static Host Sites

Then to run the proxy you can simply execute `gomost` within the root directory
containing the static resources.

By default, the current directory where the program is executed will be used
as the root static folder. If you wish to change this, you can provide a configuration
file such as:

```
  host: :8080
  StaticDir: /the/path/to/the/root/dir
```

then run `gomost -c=myconf.yaml`

### Application Proxy

If you wish to proxy requests to another application you need to provide a YAML configuration file that provides the proxy host mappings.

```
  host: :8080
  proxies:
    -
      proxy: www.dev1.com
      host: http://localhost:8090
    -
      proxy: www.dev2.com
      host: http://localhost:8091
```

then run `gomost -c=myconf.yaml`

### Embed Host Handler

You can also embed the proxy into your own application allowing you to create go application
endpoints for each host.

```go
  package main

  import (
    "os"

  	"github.com/landonia/gomost/proxy"
  )

  func main() {

  	// Initialise the server with the default config
  	p, err := proxy.Default()
  	if err != nil {
  		os.Exit(1)
  	}

    // You can also use the proxy.Setup(config) function if you have more
    // proxy configs to provide

    // Add a custom handler for a domain
  	p.AddHostHandler("www.dev1.com", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
  		// ... Handle the request
  	}))

    // You can also provide a server mux or the handler from any custom web frameworks
    sm := http.NewServeMux()
  	sm.HandleFunc("/mypath", func(resp http.ResponseWriter, req *http.Request) {
  		// ... Handle the request
  	})
    p.AddHostHandler("www.dev2.com", sm)

  	// Handle any requests
  	if err = p.Service(); err != nil {
  		os.Exit(1)
  	}
  }
```

Remember that you can use a combination of static, proxy and local handlers for each host.

### Config Options

There are multiple other configuration properties than can be provided to the program.

```
  host: :80 // The local address - Set to ':80' when in production
  loglevel: fatal|error|warn|info|debug|trace // info by default
  StaticDir: /the/path/to/the/root/dir // The location of the static resources
  proxies:
    -
      proxy: www.dev1.com
      host: http://localhost:8090
    -
      proxy: www.dev2.com
      host: http://localhost:8091
  ssl:
    enable: true // false by default
    certfile: /the/path/to/the/cert/file
    keyfile: /the/path/to/the/key/file
```

## About

gomost was written by [Landon Wainwright](http://www.landotube.com) | [GitHub](https://github.com/landonia).

Follow me on [Twitter @landotube](http://www.twitter.com/landotube)! Although I don't really tweet much tbh.
