package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mainHandler(t *testing.T) {
	type want struct {
		code      int
		response  string
		headerKey string
		headerVal string
		failed    bool
	}
	type request struct {
		url    string
		method string
		data   string
		want   want
	}
	tests := []struct {
		name     string
		requests []request
	}{
		{
			name: "positive POST&GET",
			requests: []request{
				{
					url:    "/",
					method: http.MethodPost,
					data:   "https://practicum.yandex.ru/",
					want: want{
						code:      http.StatusCreated,
						response:  "http://localhost:8080/QrPnX5IU",
						headerKey: "Content-Type",
						headerVal: "text/plain",
						failed:    false,
					},
				},
				{
					url:    "/QrPnX5IU",
					method: http.MethodGet,
					data:   "",
					want: want{
						code:      http.StatusTemporaryRedirect,
						response:  "",
						headerKey: "Location",
						headerVal: "https://practicum.yandex.ru/",
						failed:    false,
					},
				},
			},
		},
		{
			name: "positive POST, negative GET",
			requests: []request{
				{
					url:    "/",
					method: http.MethodPost,
					data:   "https://practicum.yandex.ru/",
					want: want{
						code:      http.StatusCreated,
						response:  "http://localhost:8080/QrPnX5IU",
						headerKey: "Content-Type",
						headerVal: "text/plain",
						failed:    false,
					},
				},
				{
					url:    "/wrong",
					method: http.MethodGet,
					data:   "",
					want: want{
						code:      http.StatusBadRequest,
						response:  "can't find related URL wrong in storage",
						headerKey: "Location",
						headerVal: "",
						failed:    true,
					},
				},
			},
		},
		{
			name: "negative POST: non empty path",
			requests: []request{
				{
					url:    "/notemptypath",
					method: http.MethodPost,
					data:   "https://practicum.yandex.ru/",
					want: want{
						code:     http.StatusBadRequest,
						response: "Wrong request path",
						failed:   true,
					},
				},
			},
		},
		{
			name: "negative GET: empty path",
			requests: []request{
				{
					url:    "/",
					method: http.MethodGet,
					data:   "",
					want: want{
						code:     http.StatusBadRequest,
						response: "Wrong input URL",
						failed:   true,
					},
				},
			},
		},
		{
			name: "wrong method",
			requests: []request{
				{
					url:    "/",
					method: http.MethodPut,
					data:   "",
					want: want{
						code:     http.StatusMethodNotAllowed,
						response: "Only POST requests are allowed!",
						failed:   true,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, r := range test.requests {
				t.Run(r.method, func(t *testing.T) {
					var req *http.Request
					if r.data != "" {
						req = httptest.NewRequest(r.method, r.url, strings.NewReader(r.data))
					} else {
						req = httptest.NewRequest(r.method, r.url, nil)
					}

					// создаём новый Recorder
					w := httptest.NewRecorder()
					mainHandler(w, req)

					res := w.Result()
					// проверяем код ответа
					assert.Equal(t, r.want.code, res.StatusCode)
					// получаем и проверяем тело запроса
					defer func() {
						if err := res.Body.Close(); err != nil {
							require.NoError(t, err)
						}
					}()
					resBody, err := io.ReadAll(res.Body)

					require.NoError(t, err)
					if r.want.failed {
						assert.Equal(t, r.want.response, strings.TrimSpace(string(resBody)))
					} else {
						require.Equal(t, r.want.response, string(resBody))
					}
					assert.Equal(t, r.want.headerVal, res.Header.Get(r.want.headerKey))
				})
			}
		})
	}
}
