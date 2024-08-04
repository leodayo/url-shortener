package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/leodayo/url-shortener/internal/app/config"
	"github.com/leodayo/url-shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURL(t *testing.T) {
	srv := httptest.NewServer(MainRouter())
	defer srv.Close()

	expectedBodyRxString := fmt.Sprintf("^%s/[a-z0-9]{%d}$", config.ExpandPath.String(), linkLength)
	expectedBodyRx := regexp.MustCompile(expectedBodyRxString)
	tests := []struct {
		name                 string
		requestHost          string
		requestMethod        string
		requestBody          string
		expectedCode         int
		expectedErrorMessage string
	}{
		{name: "Valid http request", requestHost: strings.Replace(srv.URL, "https", "http", 1), requestMethod: http.MethodPost, requestBody: "http://example.com", expectedCode: http.StatusCreated},
		{name: "Valid https request", requestHost: srv.URL, requestMethod: http.MethodPost, requestBody: "https://example.com", expectedCode: http.StatusCreated},
		{name: "Bad request, body contains not a URL", requestHost: srv.URL, requestMethod: http.MethodPost, requestBody: "not a URL", expectedCode: http.StatusBadRequest, expectedErrorMessage: "Invalid URL"},
		{name: "Bad request, not supported method", requestHost: srv.URL, requestMethod: http.MethodGet, requestBody: "http://example.com", expectedCode: http.StatusMethodNotAllowed, expectedErrorMessage: ""},
		{name: "Not found, wrong method", requestHost: srv.URL + "/some/deeper/path", requestMethod: http.MethodPost, requestBody: "http://example.com", expectedCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := resty.New().R()
			request.Method = tt.requestMethod
			request.URL = tt.requestHost
			request.Body = tt.requestBody

			response, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, response.StatusCode(), "expected status [%v], got [%v]", tt.expectedCode, response.StatusCode())
			if tt.expectedCode == http.StatusCreated {
				responseBody := string(response.Body())
				responseBody = strings.TrimSpace(responseBody)
				assert.Regexp(t, expectedBodyRx, responseBody, "expected body to match [%v], got [%v]", expectedBodyRxString, responseBody)
			}

			if tt.expectedErrorMessage != "" {
				responseBody := string(response.Body())
				responseBody = strings.TrimSpace(responseBody)
				assert.Equal(t, tt.expectedErrorMessage, responseBody, "expected error message [%v], got [%v]", tt.expectedErrorMessage, responseBody)
			}
		})
	}
}

func TestShortenURLJSON(t *testing.T) {
	srv := httptest.NewServer(MainRouter())
	defer srv.Close()
	endpointURL := srv.URL + "/api/shorten"

	expectedURLRxString := fmt.Sprintf("^%s/[a-z0-9]{%d}$", config.ExpandPath.String(), linkLength)
	expectedURLRx := regexp.MustCompile(expectedURLRxString)
	tests := []struct {
		name          string
		requestHost   string
		requestMethod string
		requestBody   string
		expectedCode  int
	}{

		{name: "Valid http request", requestHost: strings.Replace(endpointURL, "https", "http", 1), requestMethod: http.MethodPost, requestBody: "{\"Url\":\"http://example.com\"}", expectedCode: http.StatusCreated},
		{name: "Valid https request", requestHost: endpointURL, requestMethod: http.MethodPost, requestBody: "{\"Url\":\"https://example.com\"}", expectedCode: http.StatusCreated},
		{name: "Bad request, body contains not a URL", requestHost: endpointURL, requestMethod: http.MethodPost, requestBody: "{\"Url\":\"not a URL\"}", expectedCode: http.StatusBadRequest},
		{name: "Bad request, not supported method", requestHost: endpointURL, requestMethod: http.MethodGet, requestBody: "{\"Url\":\"http://example.com\"}", expectedCode: http.StatusMethodNotAllowed},
		{name: "Not found, wrong method", requestHost: endpointURL + "/some/deeper/path", requestMethod: http.MethodPost, requestBody: "{\"Url\":\"http://example.com\"}", expectedCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := resty.New().R()
			request.Method = tt.requestMethod
			request.URL = tt.requestHost
			request.Body = tt.requestBody
			request.Header.Add("Content-Type", "application/json")

			response, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, response.StatusCode(), "expected status [%v], got [%v]", tt.expectedCode, response.StatusCode())
			if tt.expectedCode == http.StatusCreated {
				var responseJSON models.ShortenResponse
				json.Unmarshal(response.Body(), &responseJSON)

				assert.Regexp(t, expectedURLRx, responseJSON.Result, "expected body to match [%v], got [%v]", expectedURLRxString, responseJSON.Result)
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	srv := httptest.NewServer(MainRouter())
	defer srv.Close()

	parsedServerURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	config.ExpandPath.Host = parsedServerURL.Host
	config.ExpandPath.Scheme = parsedServerURL.Scheme

	validOriginalURL := "https://example.com"

	request := resty.New().R()
	request.Method = http.MethodPost
	request.URL = srv.URL
	request.Body = validOriginalURL
	response, err := request.Send()

	require.NoError(t, err, "error making HTTP request")
	require.Equal(t, http.StatusCreated, response.StatusCode(), "Failed to create shortened URL: expected status [%v], got [%v]", http.StatusCreated, response.StatusCode())
	shortenedURL := string(response.Body())
	shortenedURL = strings.TrimSpace(shortenedURL)

	tests := []struct {
		name                 string
		requestHost          string
		requestMethod        string
		expectedCode         int
		expectedErrorMessage string
	}{
		{name: "Valid request", requestHost: shortenedURL, requestMethod: http.MethodGet, expectedCode: http.StatusTemporaryRedirect},
		{name: "Bad request, not supported method", requestHost: shortenedURL, requestMethod: http.MethodPost, expectedCode: http.StatusMethodNotAllowed, expectedErrorMessage: ""},
		{name: "Not found", requestHost: shortenedURL + "invalidId", requestMethod: http.MethodGet, expectedCode: http.StatusNotFound, expectedErrorMessage: "Link not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
				// Prevent auto redirect
				return http.ErrUseLastResponse
			}))
			request := client.R()
			request.Method = tt.requestMethod
			request.URL = tt.requestHost

			response, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, response.StatusCode(), "expected status [%v], got [%v]", tt.expectedCode, response.StatusCode())

			if tt.expectedCode == http.StatusTemporaryRedirect {
				response.Header()
				actualLocation := response.Header().Get("Location")
				assert.Equal(t, validOriginalURL, actualLocation, "Expected Location header to contain [%v], got [%v]", validOriginalURL, actualLocation)
			}

			if tt.expectedErrorMessage != "" {
				responseBody := string(response.Body())
				responseBody = strings.TrimSpace(responseBody)
				assert.Equal(t, tt.expectedErrorMessage, responseBody, "expected error message [%v], got [%v]", tt.expectedErrorMessage, responseBody)
			}
		})
	}
}
