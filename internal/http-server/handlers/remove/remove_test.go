package remove_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azizjon12/url-shortener/internal/http-server/handlers/remove"
	"github.com/azizjon12/url-shortener/internal/http-server/handlers/remove/mocks"
	"github.com/azizjon12/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name       string
		alias      string
		respStatus int
		respError  string
		mockError  error
	}{
		{
			name:       "Success",
			alias:      "test_alias",
			respStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.alias != "" {
				urlDeleterMock.
					On("DeleteURL", tc.alias).Return(tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Delete("/{alias}", remove.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/"+tc.alias, nil)
			require.NoError(t, err)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.respStatus, resp.StatusCode)
		})
	}
}
