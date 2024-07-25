package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/app/entity"
	"github.com/leodayo/url-shortener/internal/app/randstr"
	"github.com/leodayo/url-shortener/internal/app/storage"
	"github.com/leodayo/url-shortener/internal/app/storage/memory"
)

const linkLength = 6

// move memory storage initialization to another place
var memoryStorage storage.Storage[string, entity.ShortenURL] = new(memory.ShortenURLMemoryStorage)

func ShortenURL(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(response, "Not supported", http.StatusMethodNotAllowed)
		return
	}

	if request.URL.Path != "/" {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	shortID, err := randstr.RandString(linkLength)
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

	ok := memoryStorage.Store(entity.ShortenURL{ID: shortID, OriginalURL: originalURL})

	if !ok {
		// Likely a collision happened
		// TODO: handle collisions gracefully
		http.Error(response, "Something went wrong", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
	response.Header().Set("Content-Type", "text/plain")

	shortenedURL := fmt.Sprintf("%s/%s", config.ExpandPath.String(), shortID)
	response.Write([]byte(shortenedURL))
}

func GetOriginalURL(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(response, "Not supported", http.StatusMethodNotAllowed)
		return
	}

	requestedID := request.PathValue("id")
	shortenURL, ok := memoryStorage.Retrieve(requestedID)
	if !ok {
		http.Error(response, "Link not found", http.StatusNotFound)
		return
	}

	http.Redirect(response, request, shortenURL.OriginalURL, http.StatusTemporaryRedirect)
}
