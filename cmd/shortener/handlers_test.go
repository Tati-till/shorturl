package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"shorturl/internal/config"
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
			name: "positive POST&GET yandex",
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
			name: "positive POST&GET google",
			requests: []request{
				{
					url:    "/",
					method: http.MethodPost,
					data:   "https://www.google.com/",
					want: want{
						code:      http.StatusCreated,
						response:  "http://localhost:8080/0OGWoMJd",
						headerKey: "Content-Type",
						headerVal: "text/plain",
						failed:    false,
					},
				},
				{
					url:    "/0OGWoMJd",
					method: http.MethodGet,
					data:   "",
					want: want{
						code:      http.StatusTemporaryRedirect,
						response:  "",
						headerKey: "Location",
						headerVal: "https://www.google.com/",
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
						code:      http.StatusInternalServerError,
						response:  "URL not found",
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
						response: "Bad Request",
						failed:   true,
					},
				},
			},
		},
		{
			name: "negative POST: wrong URL",
			requests: []request{
				{
					url:    "/",
					method: http.MethodPost,
					data:   "://example.com",
					want: want{
						code:     http.StatusBadRequest,
						response: "Invalid URL",
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
						response: "Bad Request",
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
						code:     http.StatusBadRequest,
						response: "Bad Request",
						failed:   true,
					},
				},
			},
		},
		{
			name: "positive POST&GET JSON",
			requests: []request{
				{
					url:    "/api/shorten",
					method: http.MethodPost,
					data:   `{"url":"https://practicum.yandex.ru"}`,
					want: want{
						code:      http.StatusCreated,
						response:  `{"result":"http://localhost:8080/ipkjUVtE"}`,
						headerKey: "Content-Type",
						headerVal: "application/json",
						failed:    false,
					},
				},
				{
					url:    "/ipkjUVtE",
					method: http.MethodGet,
					data:   "",
					want: want{
						code:      http.StatusTemporaryRedirect,
						response:  "",
						headerKey: "Location",
						headerVal: "https://practicum.yandex.ru",
						failed:    false,
					},
				},
			},
		},
		{
			name: "negative POST, empty url",
			requests: []request{
				{
					url:    "/api/shorten",
					method: http.MethodPost,
					data:   `{"url":""}`,
					want: want{
						code:     http.StatusBadRequest,
						response: "Invalid URL",
						failed:   true,
					},
				},
			},
		},
		{
			name: "negative POST, empty JSON",
			requests: []request{
				{
					url:    "/api/shorten",
					method: http.MethodPost,
					data:   `{}`,
					want: want{
						code:     http.StatusBadRequest,
						response: "Invalid URL",
						failed:   true,
					},
				},
			},
		},
	}

	config.ParseFlags()
	ts := httptest.NewServer(mainRouter())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, r := range test.requests {
				t.Run(r.method, func(t *testing.T) {
					var (
						req *http.Request
						err error
					)
					if r.data != "" {
						req, err = http.NewRequest(r.method, ts.URL+r.url, strings.NewReader(r.data))
					} else {
						req, err = http.NewRequest(r.method, ts.URL+r.url, nil)
					}
					require.NoError(t, err)

					client := ts.Client()
					client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse // Prevents following redirects
					}

					res, err := client.Do(req)
					require.NoError(t, err)

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
