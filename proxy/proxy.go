// Copyright 2016 Landonia Ltd. All rights reserved.

package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/landonia/golog"
)

var (
	logger = golog.New("proxy.Proxy")
)

// Proxy is the root server
type Proxy struct {
	rs           *http.Server                      // The actual server
	vs           *http.Server                      // The virtual redirect server
	config       Configuration                     // The configuration
	handlers     map[string]http.Handler           // The local handlers
	proxies      map[string]*httputil.ReverseProxy // The proxies to the host->proxy
	proxyHandler http.Handler                      // The root proxy handler
	exit         chan error                        // When to shutdown the server
}

// Setup will initialise the proxy and must be called before any other functions
func Setup(config Configuration) (*Proxy, error) {
	gm := &Proxy{}
	gm.config = config
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
		} else if gm.config.StaticDir != "" {
			logger.Trace("Serve: %v: Path: %s", req.Host, req.URL.String())

			// Just attempt to serve the file/directory specified by the host
			http.ServeFile(resp, req, path.Join(gm.config.StaticDir, req.Host))
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

// Service will start the server and handle the requests
func (gm *Proxy) Service() (err error) {

	// Initialise the server if one has not been provided
	gm.rs = &http.Server{
		Addr:    gm.config.Addr,
		Handler: gm.proxyHandler,
	}

	// Attempt to start the service
	if gm.rs == nil {
		err = fmt.Errorf("Setup() must be called")
	} else {
		logger.Info("Starting Proxy server at address: %s", gm.config.Addr)
		gm.exit = make(chan error)

		// Launch the server
		go func() {
			gm.exit <- gm.Listen()
		}()

		// Block until we receive the exit
		err = <-gm.exit
		logger.Info("Proxy server has shutdown at address: %s", gm.config.Addr)
	}
	return
}

// Listen will create the handler using the configuration to determine whether to use
// SSL (you have to specifically disable SSL) and whether you provide your own
// cert files or to use letsencrypt to automatically get the certs (by default)
func (gm *Proxy) Listen() error {
	addr := ParseHost(gm.config.Addr)
	logger.Info("Address: %s", addr)
	var ln net.Listener
	var err error

	// If the certificates have been provided then use them otherwise
	// use the auto letsencrypt
	if gm.config.SSL.Default.CertFile != "" && gm.config.SSL.Default.KeyFile != "" {
		ln, err = TLS(addr, gm.config.SSL.Default.CertFile, gm.config.SSL.Default.KeyFile)
	} else if !gm.config.SSL.DisableLetsEncrypt {
		if gm.config.Prod {
			ln, err = LETSENCRYPTPROD(addr)
		} else {
			ln, err = LETSENCRYPT(addr)
		}
	} else {

		// Fall back to a standard listener
		ln, err = TCP4(addr)
	}
	if err != nil {
		logger.Fatal("Cannot get SSL listener: %s", err.Error())
	}

	// If we should redirect the traffic
	if gm.config.SSL.RedirectHTTP.Enable {
		realSSLPort := ""
		if i := strings.Index(gm.config.Addr, ":"); i != -1 &&
			gm.config.Addr[i+1:] != "443" {
			realSSLPort = gm.config.Addr[i+1:]
		}

		// We will need to start a second http server to redirect the traffic to
		// the SSL version
		gm.vs = &http.Server{
			Addr: gm.config.SSL.RedirectHTTP.Addr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// What is the host that has been used? We need to redirect this request
				// to the correct HTTPS URI
				realHost := r.Host
				if i := strings.Index(realHost, ":"); i != -1 {
					realHost = realHost[:i]
				}
				redirectTo := "https://" + realHost + realSSLPort + r.RequestURI
				logger.Debug("Forwarding non-SSL request %s -> https", "http://"+r.Host+r.RequestURI)
				http.Redirect(w, r, redirectTo, http.StatusMovedPermanently)
			}),
		}

		// Attempt to listen to the server
		go func() {
			logger.Info("Starting SSL forwarding server at address: %s", gm.vs.Addr)
			if err = gm.vs.ListenAndServe(); err != nil {
				logger.Fatal("Cannot get SSL listener: %s", err.Error())
			}
		}()
	}
	return gm.rs.Serve(ln)
}

// Shutdown will force the Service function to exit
func (gm *Proxy) Shutdown() {
	gm.exit <- nil
}

// Proxy not really a proxy, it's just
// starts a server listening on proxyAddr but redirects all requests to the redirectToSchemeAndHost+$path
// nothing special, use it only when you want to start a secondary server which its only work is to redirect from one requested path to another
//
// returns a close function
// func Proxy(proxyAddr string, redirectSchemeAndHost string) func() error {
// 	proxyAddr = ParseHost(proxyAddr)
//
// 	// override the handler and redirect all requests to this addr
// 	h := ProxyHandler(proxyAddr, redirectSchemeAndHost)
// 	prx := New(OptionDisableBanner(true))
// 	prx.Router = h
//
// 	go prx.Listen(proxyAddr)
//
// 	return prx.Close
// }
