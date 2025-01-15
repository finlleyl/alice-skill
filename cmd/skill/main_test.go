package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhook(t *testing.T) {
	handler := http.HandlerFunc(webhook)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	successBody := `{
		"response": {
		"text": "Извините, я пока ничего не умею"
		},
		"version: "1.0"
	}`

	testCases := []struct {
		method       string
		expectedCode int
		expectedBody string
	}{
		{method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
		{method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
		{method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},
		{method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: successBody},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn`t match expected")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}

}
