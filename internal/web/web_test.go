package web_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"

	"github.com/igkostyuk/yalantis/internal/web"
	"github.com/pkg/errors"
)

func TestApp(t *testing.T) {
	t.Run("it correct process on GET", func(t *testing.T) {
		stubShutdown := make(chan os.Signal, 1)
		want := "test"
		app := web.NewApp(stubShutdown, SrubMiddleware(want, nil))
		app.Handle(http.MethodGet, "/", StubHandler)

		request := newGetRequest()
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)
		got := response.Body.String()

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
		select {
		case <-stubShutdown:
			t.Errorf("channel was closed")
		default:
		}
	})
	t.Run("it should send signal on error", func(t *testing.T) {
		stubShutdown := make(chan os.Signal, 1)
		want := "test"
		app := web.NewApp(stubShutdown, SrubMiddleware(want, errors.New("test error")))
		app.Handle(http.MethodGet, "/", StubHandler)

		request := newGetRequest()
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		select {
		case sig := <-stubShutdown:
			if sig != syscall.SIGTERM {
				t.Errorf("signal must be SIGTERM")
			}
		default:
			t.Errorf("channel was NOT closed")
		}
	})
}
func TesMiddleware(t *testing.T) {
	tests := []struct {
		name string
		mw   []string
		want string
	}{
		{
			name: "first middleware",
			mw:   []string{"1", "", ""},
			want: "1",
		},
		{
			name: "second middleware",
			mw:   []string{"", "2", ""},
			want: "2",
		},
		{
			name: "first middleware",
			mw:   []string{"", "", "3"},
			want: "3",
		},
	}

	stubShutdown := make(chan os.Signal, 1)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			app := web.NewApp(stubShutdown, SrubMiddleware(tt.mw[1], nil), SrubMiddleware(tt.mw[2], nil), SrubMiddleware(tt.mw[3], nil))
			app.Handle(http.MethodGet, "/", StubHandler)

			request := newGetRequest()
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)
			got := response.Body.String()

			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

func newGetRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	return req
}

func SrubMiddleware(body string, err error) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			_, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}
			if body != "" {
				fmt.Fprint(w, body)
			}
			return err
		}
		return h
	}
	return m
}

func StubHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}
