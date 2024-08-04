package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/app/entity"
	"github.com/leodayo/url-shortener/internal/app/randstr"
	"github.com/leodayo/url-shortener/internal/app/storage"
	"github.com/leodayo/url-shortener/internal/app/storage/memory"
	"github.com/leodayo/url-shortener/internal/logger"
	"github.com/leodayo/url-shortener/internal/models"
	"go.uber.org/zap"
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

func ShortenURLJSON(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(response, "Not supported", http.StatusMethodNotAllowed)
		return
	}

	if request.Header.Get("Content-Type") != "application/json" {
		http.Error(response, "Content-Type not supported", http.StatusBadRequest)
		return
	}

	var shortenRequest models.ShortenRequest
	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&shortenRequest); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	originalURL := shortenRequest.Url
	parsedURL, err := url.Parse(originalURL)
	if err != nil || parsedURL.Host == "" {
		http.Error(response, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortID, err := randstr.RandString(linkLength)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
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
	response.Header().Set("Content-Type", "application/json")

	shortenedURL := fmt.Sprintf("%s/%s", config.ExpandPath.String(), shortID)
	shortenResponse := models.ShortenResponse{
		Result: shortenedURL,
	}

	enc := json.NewEncoder(response)
	if err := enc.Encode(shortenResponse); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
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
