package config

import (
	"flag"
	"net/url"
	"os"
)

var (
	ServerAddress   string
	ExpandPath      url.URL
	FileStoragePath string
)

func init() {
	ServerAddress = "localhost:8080"
	defaultExpandPath, _ := url.Parse("http://localhost:8080/expand")
	ExpandPath = *defaultExpandPath
	FileStoragePath = "storage.json"
}

func ParseFlags() {
	flag.StringVar(&ServerAddress, "a", ServerAddress, "server address")
	flag.Func("b", "base route to expand shortened URL", parseExpandPathFlag)
	flag.StringVar(&FileStoragePath, "f", FileStoragePath, "file storage path")

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

	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		FileStoragePath = fileStoragePath
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
