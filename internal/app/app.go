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
	mux.HandleFunc("/", shortenURL)
	mux.HandleFunc("/{id}", getOriginalURL)

	memoryStorage = new(memory.ShortenURLMemoryStorage)

	return http.ListenAndServe("localhost:8080", mux)
}

func shortenURL(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(response, "Not supported", http.StatusBadRequest)
		return
	}

	if request.URL.Path != "/" {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	shortURL, err := randstr.RandString(linkLength)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, "Something went wrong", http.StatusInternalServerError)
		return
	}

	originalURL := string(body)
	parsedURL, err := url.Parse(originalURL)
	if err != nil || parsedURL.Host == "" {
		http.Error(response, "Invalid URL", http.StatusBadRequest)
		return
	}

	ok := memoryStorage.Store(entity.ShortenURL{URL: shortURL, OriginalURL: originalURL})

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
	response.Write([]byte(prefix + shortURL))
}

func getOriginalURL(response http.ResponseWriter, request *http.Request) {
	requestedID := request.PathValue("id")
	shortenURL, ok := memoryStorage.Retrieve(requestedID)
	if !ok {
		http.Error(response, "Link not found", http.StatusBadRequest)
		return
	}

	http.Redirect(response, request, shortenURL.OriginalURL, http.StatusTemporaryRedirect)
}
