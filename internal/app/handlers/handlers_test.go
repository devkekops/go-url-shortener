package handlers_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devkekops/go-url-shortener/internal/app/handlers"
	"github.com/devkekops/go-url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	//disable autoredirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	respBodyStr := strings.TrimSuffix(string(respBody), "\n")

	return resp, respBodyStr
}

func TestServer(t *testing.T) {
	linkRepo := storage.NewLinkRepoMemory()
	s := handlers.NewBaseHandler(linkRepo, "http://localhost:8080", "secret")

	ts := httptest.NewServer(s)
	defer ts.Close()

	type want struct {
		code        int
		contentType string
		body        string
		location    string
	}
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   want
	}{
		{
			name:   "POST / positive",
			method: "POST",
			path:   "/",
			body:   "https://yandex.ru",
			want: want{
				code:        201,
				contentType: "text/plain; charset=utf-8",
				body:        "http://localhost:8080/1",
			},
		},
		{
			name:   "POST / incorrect URL",
			method: "POST",
			path:   "/",
			body:   "http/yandexru",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				body:        "URL is incorrect",
			},
		},
		{
			name:   "GET /1 positive",
			method: "GET",
			path:   "/1",
			want: want{
				code:     307,
				location: "https://yandex.ru",
			},
		},
		{
			name:   "GET /abc2 non-existent link",
			method: "GET",
			path:   "/abc2",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
				body:        "Not found",
			},
		},
		{
			name:   "GET /$ incorrect request",
			method: "GET",
			path:   "/$",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				body:        "Bad request",
			},
		},
		{
			name:   "PUT / request",
			method: "PUT",
			path:   "/",
			body:   "put test body",
			want: want{
				code: 405,
			},
		},
		{
			name:   "POST /api/shorten positive",
			method: "POST",
			path:   "/api/shorten",
			body:   `{"url":"https://sberbank.ru"}`,
			want: want{
				code:        201,
				contentType: "application/json",
				body:        `{"result":"http://localhost:8080/2"}`,
			},
		},
		{
			name:   "POST /api/shorten incorrect JSON",
			method: "POST",
			path:   "/api/shorten",
			body:   `{"url":https://sberbank.ru}`,
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				body:        "Bad request",
			},
		},
		{
			name:   "POST /api/shorten incorrect URL",
			method: "POST",
			path:   "/api/shorten",
			body:   `{"url":"http/sberbankru"}`,
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				body:        "URL is incorrect",
			},
		},
		{
			name:   "GET /api/user/urls no content",
			method: "GET",
			path:   "/api/user/urls",
			want: want{
				code:        204,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.path, tt.body)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
			assert.Equal(t, tt.want.body, body)
		})
	}
}
