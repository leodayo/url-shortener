package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURL(t *testing.T) {
	validHttpsHost := "https://somehost/"
	expectedBodyRxString := fmt.Sprintf("^(http|https)%s[a-z0-9]{%d}$", strings.Replace(validHttpsHost, "https", "", 1), linkLength)
	expectedBodyRx := regexp.MustCompile(expectedBodyRxString)
	tests := []struct {
		name                 string
		requestHost          string
		requestMethod        string
		requestBody          string
		expectedCode         int
		expectedErrorMessage string
	}{
		{name: "Valid http request", requestHost: strings.Replace(validHttpsHost, "https", "http", 1), requestMethod: http.MethodPost, requestBody: "http://example.com", expectedCode: http.StatusCreated},
		{name: "Valid https request", requestHost: validHttpsHost, requestMethod: http.MethodPost, requestBody: "https://example.com", expectedCode: http.StatusCreated},
		{name: "Bad request, body contains not a URL", requestHost: validHttpsHost, requestMethod: http.MethodPost, requestBody: "not a URL", expectedCode: http.StatusBadRequest, expectedErrorMessage: "Invalid URL"},
		{name: "Bad request, not supported method", requestHost: validHttpsHost, requestMethod: http.MethodGet, requestBody: "http://example.com", expectedCode: http.StatusBadRequest, expectedErrorMessage: "Not supported"},
		{name: "Not found, wrong method", requestHost: validHttpsHost + "/some/deeper/path", requestMethod: http.MethodPost, requestBody: "http://example.com", expectedCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.requestMethod, tt.requestHost, strings.NewReader(tt.requestBody))
			response := httptest.NewRecorder()
			defer response.Result().Body.Close()

			ShortenURL(response, request)

			assert.Equal(t, tt.expectedCode, response.Code, "expected status [%v], got [%v]", tt.expectedCode, response.Code)
			if tt.expectedCode == http.StatusCreated {
				responseBody := response.Body.String()
				responseBody = strings.TrimSpace(responseBody)
				assert.Regexp(t, expectedBodyRx, responseBody, "expected body to match [%v], got [%v]", expectedBodyRxString, responseBody)
			}

			if tt.expectedErrorMessage != "" {
				responseBody := response.Body.String()
				responseBody = strings.TrimSpace(responseBody)
				assert.Equal(t, tt.expectedErrorMessage, responseBody, "expected error message [%v], got [%v]", tt.expectedErrorMessage, responseBody)
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	validHost := "https://somehost/"
	validOriginalUrl := "https://example.com"

	request := httptest.NewRequest(http.MethodPost, validHost, strings.NewReader(validOriginalUrl))
	response := httptest.NewRecorder()
	defer response.Result().Body.Close()

	ShortenURL(response, request)
	require.Equal(t, http.StatusCreated, response.Code, "Failed to create shortened URL: expected status [%v], got [%v]", http.StatusCreated, response.Code)
	shortenedURL := response.Body.String()
	shortenedURL = strings.TrimSpace(shortenedURL)

	tests := []struct {
		name                 string
		requestHost          string
		requestMethod        string
		expectedCode         int
		expectedErrorMessage string
	}{
		// {name: "Valid request", requestHost: shortenedURL, requestMethod: http.MethodGet, expectedCode: http.StatusTemporaryRedirect},
		// {name: "Bad request, not supported method", requestHost: shortenedURL, requestMethod: http.MethodPost, expectedCode: http.StatusBadRequest, expectedErrorMessage: "Not supported"},
		{name: "Not found", requestHost: validHost + "invalidId", requestMethod: http.MethodGet, expectedCode: http.StatusBadRequest, expectedErrorMessage: "Link not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.requestMethod, tt.requestHost, nil)
			shortenSplit := strings.Split(tt.requestHost, "/")
			requestedId := shortenSplit[len(shortenSplit)-1]
			request.SetPathValue("id", requestedId)
			response := httptest.NewRecorder()
			defer response.Result().Body.Close()

			GetOriginalURL(response, request)
			assert.Equal(t, tt.expectedCode, response.Code, "expected status [%v], got [%v]", tt.expectedCode, response.Code)

			if tt.expectedCode == http.StatusTemporaryRedirect {
				response.Header()
				actualLocation := response.Header().Get("Location")
				assert.Equal(t, validOriginalUrl, actualLocation, "Expected Location header to contain [%v], got [%v]", validOriginalUrl, actualLocation)
			}

			if tt.expectedErrorMessage != "" {
				responseBody := response.Body.String()
				responseBody = strings.TrimSpace(responseBody)
				assert.Equal(t, tt.expectedErrorMessage, responseBody, "expected error message [%v], got [%v]", tt.expectedErrorMessage, responseBody)
			}
		})
	}
}
