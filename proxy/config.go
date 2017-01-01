// Copyright 2016 Landonia Ltd. All rights reserved.

package proxy

import (
	"bytes"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Configuration wraps the settings required for the app
type Configuration struct {
	Prod      bool         `yaml:"prod"`     // Whether in production (this will change the SSL handler)
	Addr      string       `yaml:"addr"`     // The host to locally bind
	LogLevel  string       `yaml:"loglevel"` // The log level to use
	StaticDir string       `yaml:"static"`   // The static hosts root directory
	Proxies   []HostConfig `yaml:"proxies"`  // The proxy information
	SSL       struct {
		RedirectHTTP struct {
			Enable bool   `yaml:"enable"` // If true this will setup a second server to redirect HTTP -> HTTPS
			Addr   string `yaml:"addr"`   // The address of the redirect
		} `yaml:"redirecthttp"`
		DisableLetsEncrypt bool `yaml:"disableletsencrypt"` // True if LetsEncrypt auto SSL should not be used
		Default            struct {
			CertFile string `yaml:"certfile"` // The certfile path
			KeyFile  string `yaml:"keyfile"`  // The keyfile path
		} `yaml:"files"`
	} `yaml:"ssl"` // The ssl information
}

// HostConfig information
type HostConfig struct {
	Proxy string `yaml:"proxy"`
	Host  string `yaml:"host"`
}

// DefaultConfig will return a sensible default configuration
func DefaultConfig() Configuration {
	conf := Configuration{}
	conf.Prod = true
	conf.Addr = DefaultSSLAddr
	conf.StaticDir = "."
	conf.LogLevel = "DEBUG"
	conf.SSL.RedirectHTTP.Enable = true
	conf.SSL.RedirectHTTP.Addr = ":80"
	return conf
}

// ParseFileConfig will return a new Configuration
func ParseFileConfig(path string) (Configuration, error) {

	// try opening the file to see if it exists
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return Configuration{}, err
	}
	conf := Configuration{}
	var b bytes.Buffer
	_, err = b.ReadFrom(file)
	if err == nil {
		err = yaml.Unmarshal(b.Bytes(), &conf)
	}
	return conf, err
}
