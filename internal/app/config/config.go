package config

import (
	"flag"
	"net/url"
	"os"
)

var (
	ServerAddress string
	ExpandPath    url.URL
)

func init() {
	ServerAddress = "localhost:8080"
	defaultExpandPath, _ := url.Parse("http://localhost:8080/expand")
	ExpandPath = *defaultExpandPath
}

func ParseFlags() {
	flag.StringVar(&ServerAddress, "a", ServerAddress, "server address")
	flag.Func("b", "base route to expand shortened URL", parseExpandPathFlag)

	flag.Parse()
}

func ParseEnv() error {
	if serverAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		ServerAddress = serverAddress
	}

	if baseURL, ok := os.LookupEnv("BASE_URL"); ok {
		parsedBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		ExpandPath = *parsedBaseURL
	}

	return nil
}

func parseExpandPathFlag(expandPath string) error {
	parsedExpandURL, err := url.Parse(expandPath)

	if err != nil {
		return err
	}

	ExpandPath = *parsedExpandURL
	return nil
}
