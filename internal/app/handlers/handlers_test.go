package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURL(t *testing.T) {
	srv := httptest.NewServer(MainRouter())
	defer srv.Close()

	expectedBodyRxString := fmt.Sprintf("^(http|https)%s/[a-z0-9]{%d}$", strings.Replace(srv.URL, "http", "", 1), linkLength)
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

func TestGetOriginalURL(t *testing.T) {
	srv := httptest.NewServer(MainRouter())
	defer srv.Close()

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
		{name: "Not found", requestHost: srv.URL + "/invalidId", requestMethod: http.MethodGet, expectedCode: http.StatusNotFound, expectedErrorMessage: "Link not found"},
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
