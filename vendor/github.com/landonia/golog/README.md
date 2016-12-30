# golog

A wrapper for the go log providing namespaces and standard levels

## Overview

The standard go log package provides the core writing methods but does
not provide any colouring or level functions. This simply provides those features.
You can overwrite the standard log on the package level if you require any
different settings.

## Installation

With a healthy Go Language installed, simply run `go get github.com/landonia/golog`

## Example
```go
  	package main

	import (
		"flag"
		"github.com/landonia/golog"
	)

	var (
		log := golog.New("mynamespace")
	)

	func main() {
		// Setup application.....
		log.Info("Application has started successfully..")

		// .. something goes wrong
		log.Error("Whoops")
	}
```

## Out of Box Example

simply run `go run $GOPATH/src/github.com/landonia/golog/cmd/example.go`

You should see output to the following:

![Example output](cmd/example.png?raw=true)

## About

golog was written by [Landon Wainwright](http://www.landotube.com) | [GitHub](https://github.com/landonia).

Follow me on [Twitter @landotube](http://www.twitter.com/landotube)! Although I don't really tweet much tbh.
