// Copyright 2016 Landonia Ltd. All rights reserved.

package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/iris-contrib/letsencrypt"
	"golang.org/x/crypto/acme/autocert"
)

// Default values for base Server conf
const (
	// DefaultSSLAddr return the default SSL addr to bind
	DefaultSSLAddr = ":443"
	// DefaultServerHostname returns the default hostname which is 0.0.0.0
	DefaultServerHostname = "0.0.0.0"
	// DefaultServerPort returns the default port which is 8080, not used
	DefaultServerPort = 8080
)

// FormatError will allow an error message to be formatted directly
type FormatError struct {
	s string
}

func (fe *FormatError) Error() string {
	return fe.s
}

// Format returns a formatted new error based on the arguments
// it does NOT change the original error's message
func (fe *FormatError) Format(a ...interface{}) error {
	return fmt.Errorf(fe.s, a...)
}

var (
	// DefaultServerAddr the default server addr which is: 0.0.0.0:8080
	DefaultServerAddr = DefaultServerHostname + ":" + strconv.Itoa(DefaultServerPort)
	errCertKeyMissing = errors.New("You should provide certFile and keyFile for TLS/SSL")
	errParseTLS       = &FormatError{"Couldn't load TLS, certFile=%q, keyFile=%q. Trace: %s"}
)

// TLS returns a new TLS Listener
func TLS(addr, certFile, keyFile string) (net.Listener, error) {

	if certFile == "" || keyFile == "" {
		return nil, errCertKeyMissing
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errParseTLS.Format(certFile, keyFile, err)
	}

	return CERT(addr, cert)
}

// CERT returns a listener which contans tls.Config with the provided certificate, use for ssl
func CERT(addr string, cert tls.Certificate) (net.Listener, error) {
	ln, err := TCP4(addr)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
	}
	return tls.NewListener(ln, tlsConfig), nil
}

// LETSENCRYPT returns a new Automatic TLS Listener using letsencrypt.org service
// receives two parameters, the first is the domain of the server
// and the second is optionally, the cache file, if you skip it then the cache directory is "./letsencrypt.cache"
// if you want to disable cache file then simple give it a value of empty string ""
//
// supports localhost domains for testing,
// but I recommend you to use the LETSENCRYPTPROD if you gonna to use it on production
func LETSENCRYPT(addr string) (net.Listener, error) {
	if portIdx := strings.IndexByte(addr, ':'); portIdx == -1 {
		addr += DefaultSSLAddr
	}

	ln, err := TCP4(addr)
	if err != nil {
		return nil, err
	}

	var m letsencrypt.Manager
	if err = m.CacheFile("./letsencrypt.cache"); err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{GetCertificate: m.GetCertificate}
	tlsLn := tls.NewListener(ln, tlsConfig)
	return tlsLn, nil
}

// LETSENCRYPTPROD returns a new Automatic TLS Listener using letsencrypt.org service
// receives two parameters, the first is the domain of the server
// and the second is optionally, the cache directory, if you skip it then the cache directory is "./certcache"
// if you want to disable cache directory then simple give it a value of empty string ""
//
// does NOT supports localhost domains for testing, use LETSENCRYPT instead.
//
// this is the recommended function to use when you're ready for production state
func LETSENCRYPTPROD(addr string) (net.Listener, error) {
	if portIdx := strings.IndexByte(addr, ':'); portIdx == -1 {
		addr += DefaultSSLAddr
	}

	ln, err := TCP4(addr)
	if err != nil {
		return nil, err
	}

	m := autocert.Manager{
		Prompt: autocert.AcceptTOS,
	} // HostPolicy is missing, if user wants it, then she/he should manually
	// configure the autocertmanager and use the `iris.Serve` to pass that listener

	m.Cache = autocert.DirCache("./certcache")
	tlsConfig := &tls.Config{GetCertificate: m.GetCertificate}
	tlsLn := tls.NewListener(ln, tlsConfig)
	return tlsLn, nil
}

// TCP4 returns a new tcp4 Listener
// *tcp6 has some bugs in some operating systems, as reported by Go Community*
func TCP4(addr string) (net.Listener, error) {
	return net.Listen("tcp4", ParseHost(addr))
}

// ParseHost tries to convert a given string to an address which is compatible with net.Listener and server
func ParseHost(addr string) string {
	// check if addr has :port, if not do it +:80 ,we need the hostname for many cases
	a := addr
	if a == "" {
		// check for os environments
		if oshost := os.Getenv("ADDR"); oshost != "" {
			a = oshost
		} else if oshost := os.Getenv("HOST"); oshost != "" {
			a = oshost
		} else if oshost := os.Getenv("HOSTNAME"); oshost != "" {
			a = oshost
			// check for port also here
			if osport := os.Getenv("PORT"); osport != "" {
				a += ":" + osport
			}
		} else if osport := os.Getenv("PORT"); osport != "" {
			a = ":" + osport
		} else {
			a = DefaultServerAddr
		}
	}
	if portIdx := strings.IndexByte(a, ':'); portIdx == 0 {
		if a[portIdx:] == ":https" {
			a = DefaultServerHostname + ":443"
		} else {
			// if contains only :port	,then the : is the first letter, so we dont have setted a hostname, lets set it
			a = DefaultServerHostname + a
		}
	}

	return a
}
