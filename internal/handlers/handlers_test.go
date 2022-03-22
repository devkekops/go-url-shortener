package handlers_test

import (
	"fmt"
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
	url := mr.idToLinkMap[id]

	fmt.Printf("FindById %d url %s\n", id, url)
	fmt.Println(mr.idToLinkMap)

	return url, nil
}

func (mr *MockedLinkRepo) Save(link string) (int64, error) {
	index := len(mr.idToLinkMap) + 1
	mr.idToLinkMap[int64(index)] = link

	fmt.Printf("Save %s with index %d\n", link, index)
	fmt.Println(mr.idToLinkMap)

	return int64(index), nil
}

func TestRootHandler(t *testing.T) {
	idToLinkMap := make(map[int64]string)
	linkRepo := NewMockedLinkRepo(idToLinkMap)
	bh := handlers.NewBaseHandler(linkRepo)

	type want struct {
		code        int
		contentType string
		body        string
	}
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   want
	}{
		{
			name:   "positive POST test #1",
			method: "POST",
			path:   "/",
			body:   "https://yandex.ru",
			want: want{
				code:        201,
				contentType: "text/plain; charset=utf-8",
				body:        "http://localhost:8080/1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(bh.RootHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			bodyBytes, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)
			body := string(bodyBytes)
			assert.Equal(t, tt.want.body, body)
		})
	}
}
