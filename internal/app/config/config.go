package config

import (
	"flag"
	"net/url"
)

var (
	ServerAddress string
	ExpandPath    url.URL
)

func init() {
	ServerAddress = "localhost:8080"
	defaultExpandPath, _ := url.Parse("http://localhost:8000/expand")
	ExpandPath = *defaultExpandPath
}

func ParseFlags() {
	flag.StringVar(&ServerAddress, "a", "localhost:8080", "server address")
	flag.Func("b", "base route to expand shortened URL", parseExpandPathFlag)

	flag.Parse()
}

func parseExpandPathFlag(expandPath string) error {
	parsedExpandUrl, err := url.Parse(expandPath)

	if err != nil {
		return err
	}

	ExpandPath = *parsedExpandUrl

	return nil
}
