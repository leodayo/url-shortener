package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/leodayo/url-shortener/internal/app/entity"
	"github.com/leodayo/url-shortener/internal/app/randstr"
	"github.com/leodayo/url-shortener/internal/app/storage"
	"github.com/leodayo/url-shortener/internal/app/storage/memory"
)

const linkLength = 6

var memoryStorage storage.Storage[string, entity.ShortenURL]

func Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", shortenUrl)
	mux.HandleFunc("/{id}", getOriginalUrl)

	memoryStorage = new(memory.ShortenURLMemoryStorage)

	return http.ListenAndServe("localhost:8080", mux)
}

func shortenUrl(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(response, "Not supported", http.StatusBadRequest)
		return
	}

	if request.URL.Path != "/" {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	shortUrl, err := randstr.RandString(linkLength)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, "Something went wrong", http.StatusInternalServerError)
		return
	}

	originalUrl := string(body)
	parsedUrl, err := url.Parse(originalUrl)
	if err != nil || parsedUrl.Host == "" {
		http.Error(response, "Invalid URL", http.StatusBadRequest)
		return
	}

	ok := memoryStorage.Store(entity.ShortenURL{URL: shortUrl, OriginalURL: originalUrl})

	if !ok {
		// Likely a collision happened
		// TODO: handle collisions gracefully
		http.Error(response, "Something went wrong", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
	response.Header().Set("Content-Type", "text/plain")

	prefix := fmt.Sprintf("%s://%s/", request.URL.Scheme, request.Host)
	if request.TLS == nil {
		prefix = "http" + prefix
	}
	response.Write([]byte(prefix + shortUrl))
}

func getOriginalUrl(response http.ResponseWriter, request *http.Request) {
	requestedId := request.PathValue("id")
	shortenUrl, ok := memoryStorage.Retrieve(requestedId)
	if !ok {
		http.Error(response, "Link not found", http.StatusBadRequest)
		return
	}

	http.Redirect(response, request, shortenUrl.OriginalURL, http.StatusTemporaryRedirect)
}
