package handlers_test

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devkekops/go-url-shortener/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockedLinkRepo struct {
	idToLinkMap map[int64]string
}

func NewMockedLinkRepo(idToLinkMap map[int64]string) *MockedLinkRepo {
	return &MockedLinkRepo{
		idToLinkMap: idToLinkMap,
	}
}

func (mr *MockedLinkRepo) FindByID(id int64) (string, error) {
	url, exist := mr.idToLinkMap[id]
	if exist == false {
		err := sql.ErrNoRows
		return "", err
	}
	//fmt.Printf("FindById %d url %s\n", id, url)
	//fmt.Println(mr.idToLinkMap)

	return url, nil
}

func (mr *MockedLinkRepo) Save(link string) (int64, error) {
	index := len(mr.idToLinkMap) + 1
	mr.idToLinkMap[int64(index)] = link
	//fmt.Printf("Save %s with index %d\n", link, index)
	//fmt.Println(mr.idToLinkMap)

	return int64(index), nil
}

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

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestServer(t *testing.T) {
	idToLinkMap := make(map[int64]string)
	linkRepo := NewMockedLinkRepo(idToLinkMap)
	s := handlers.NewServer(linkRepo)

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
			name:   "positive POST test",
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
			name:   "positive GET test",
			method: "GET",
			path:   "/1",
			want: want{
				code:     307,
				location: "https://yandex.ru",
			},
		},
		{
			name:   "GET non-existent link",
			method: "GET",
			path:   "/abc2",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
				body:        "Not found\n",
			},
		},
		{
			name:   "PUT request",
			method: "PUT",
			path:   "/",
			body:   "put test body",
			want: want{
				code: 405,
			},
		},
		{
			name:   "GET incorrect request #1",
			method: "GET",
			path:   "/$",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				body:        "Bad request\n",
			},
		},
		{
			name:   "POST non-URL string",
			method: "POST",
			path:   "/",
			body:   "http/yandexru",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
				body:        "URL is incorrect\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.path, tt.body)

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
			assert.Equal(t, tt.want.body, body)
		})
	}
}
