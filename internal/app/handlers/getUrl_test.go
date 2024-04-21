package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GTedya/shortener/config"
	"github.com/GTedya/shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetURLByID(t *testing.T) {
	data := make(map[string]string)
	data["testID"] = "http://localhost:8080/testID"

	type args struct {
		url         string
		method      string
		contentType string
	}

	type want struct {
		code        int
		location    string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "pos test",
			want: want{
				code:        307,
				contentType: "text/plain; application/json",
				location:    data["testID"],
			},
			args: args{
				url:         "http://localhost:8080/testID",
				method:      http.MethodGet,
				contentType: "text/plain; application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := chi.NewRouter()
			conf := config.Config{Address: "localhost:8080", URL: "short"}
			log := &zap.SugaredLogger{}

			store, err := storage.NewStore(conf, nil)
			if err != nil {
				t.Log(err)
			}

			ctx := context.TODO()
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			h := handler{log: log, conf: conf, store: store}
			err = h.store.SaveURL(ctx, "", "http://localhost:8080/testID", "testID")
			if err != nil {
				t.Log(err)
			}

			r.Get("/{id:[a-zA-Z0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
				h.getURLByID(writer, request)
			})

			req := httptest.NewRequest(http.MethodGet, "/testID", nil)
			recorder := httptest.NewRecorder()

			r.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Log(fmt.Errorf("response body closing error: %w", err))
				}
			}()

			assert.Equal(t, test.want.code, res.StatusCode)

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("location"))
		})
	}
}

func BenchmarkGetURLByID(b *testing.B) {
	r := chi.NewRouter()
	conf := config.Config{Address: "localhost:8080", URL: "short"}
	log := &zap.SugaredLogger{}
	store, err := storage.NewStore(conf, nil)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.TODO()

	h := handler{log: log, conf: conf, store: store}
	err = h.store.SaveURL(ctx, "", "http://localhost:8080/testID", "testID")
	if err != nil {
		b.Fatal(err)
	}

	r.Get("/{id:[a-zA-Z0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
		h.getURLByID(writer, request)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/testID", nil)
		recorder := httptest.NewRecorder()

		r.ServeHTTP(recorder, req)
	}
}
