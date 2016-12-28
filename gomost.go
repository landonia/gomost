// Copyright 2016 Landonia Ltd. All rights reserved.

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/landonia/golog"
	"github.com/landonia/gomost/proxy"
)

var (
	logger = golog.New("gomost.Main")
)

// bootstrap the application
func main() {

	// Get access to the settings
	configPath := flag.String("c", "", "The configuration file")
	flag.Parse()
	var config proxy.Configuration
	var err error
	if *configPath != "" {

		// parse the config if it is available
		config, err = proxy.ParseFileConfig(*configPath)
	} else {

		// otherwise create a basic config that will host the static files from
		// the current directory
		config = proxy.Configuration{
			StaticDir: ".",
		}
	}
	if err != nil {
		logger.Fatal("Could not parse configuration: %s", err.Error())
	}

	// Default the local host bind address
	if config.Host == "" {
		config.Host = ":8080"
	}
	golog.LogLevel(config.LogLevel)

	// initialise the server
	p, err := proxy.Setup(config)
	if err != nil {
		logger.Fatal("Could not start Gomost server: %s", err.Error())
	}

	// Wait for a shutdown signal
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			p.Shutdown()
		}()
	}()

	// Handle any requests
	if err = p.Service(); err != nil {
		logger.Fatal("Error shutting down Gomost server: %s", err.Error())
	}
}
