// Copyright 2016 Landonia Ltd. All rights reserved.

package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/landonia/golog"
)

var (
	logger = golog.New("proxy.Proxy")
)

// Proxy is the root server
type Proxy struct {
	Configuration                                   // The configuration
	*http.Server                                    // The actual server
	handlers      map[string]http.Handler           // The local handlers
	proxies       map[string]*httputil.ReverseProxy // The proxies to the host->proxy
	proxyHandler  http.Handler                      // The root proxy handler
	exit          chan error                        // When to shutdown the server
}

// Default will initialise the proxy with default settings
func Default() (*Proxy, error) {
	return Setup(Configuration{Host: ":8080", StaticDir: "."})
}

// Setup will initialise the proxy and must be called before any other functions
func Setup(config Configuration) (*Proxy, error) {
	gm := &Proxy{}
	gm.Configuration = config
	gm.handlers = make(map[string]http.Handler)
	gm.proxies = make(map[string]*httputil.ReverseProxy)

	// If there are any proxies then we need to set them up as well
	for _, proxy := range config.Proxies {
		if u, err := url.Parse(proxy.Host); err == nil {
			gm.proxies[proxy.Proxy] = httputil.NewSingleHostReverseProxy(u)
		} else {
			logger.Warn("Could not parse Host: %s", err.Error())
		}
	}

	// Create the root handler
	gm.proxyHandler = http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		// We need to extract the host header and then forward to the correct handler
		if handler, hExists := gm.handlers[req.Host]; hExists {
			logger.Trace("Handler: %v: Path: %s", req.Host, req.URL.String())

			// Forward to the local handler
			handler.ServeHTTP(resp, req)
		} else if proxy, pExists := gm.proxies[req.Host]; pExists {
			logger.Trace("Proxy: %v: Path: %s", req.Host, req.URL.String())

			// Forward to the proxy
			proxy.ServeHTTP(resp, req)
		} else if gm.StaticDir != "" {
			logger.Trace("Serve: %v: Path: %s", req.Host, req.URL.String())

			// Just attempt to serve the file/directory specified by the host
			http.ServeFile(resp, req, path.Join(gm.StaticDir, req.Host))
		} else {
			logger.Trace("Serve: %v: Notfound: %s", req.Host, req.URL.String())
			resp.WriteHeader(http.StatusNotFound)
		}
	})
	return gm, nil
}

// AddHostHandler will add the handler that will be used for the specified
// host allowing you to run a Go application within the proxy
func (gm *Proxy) AddHostHandler(host string, handler http.Handler) error {
	if host == "" {
		return fmt.Errorf("The host cannot be empty")
	}
	if gm.handlers == nil {
		return fmt.Errorf("Setup() must be called")
	}
	gm.handlers[host] = handler
	return nil
}

// SetServer will overide the default Server that will be used
func (gm *Proxy) SetServer(server *http.Server) {
	gm.Server = server
}

// Service will start the server and handle the requests
func (gm *Proxy) Service() (err error) {

	// Initialise the server if one has not been provided
	if gm.Server == nil {
		gm.Server = &http.Server{
			Addr: gm.Host,
		}
	}

	// Update the handler to the proxy one
	gm.Server.Handler = gm.proxyHandler

	// Attempt to start the service
	if gm.Server == nil {
		err = fmt.Errorf("Setup() must be called")
	} else {
		logger.Info("Starting Proxy server at address: %s", gm.Addr)
		gm.exit = make(chan error)

		// Launch the server
		go func() {
			if gm.SSL.Enable {
				gm.exit <- gm.ListenAndServeTLS(gm.SSL.CertFile, gm.SSL.KeyFile)
			} else {
				gm.exit <- gm.ListenAndServe()
			}
		}()

		// Block until we receive the exit
		err = <-gm.exit
		logger.Info("Proxy server has shutdown at address: %s", gm.Addr)
	}
	return
}

// Shutdown will force the Service function to exit
func (gm *Proxy) Shutdown() {
	gm.exit <- nil
}
