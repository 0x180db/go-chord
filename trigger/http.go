package trigger

import (
	"context"
	"net/http"

	"github.com/0x180db/go-chord"
	"github.com/0x180db/go-conduit"
)

type HttpContext struct {
	Writer  http.ResponseWriter
	Request *http.Request
	done    chan struct{}
}

func (h HttpContext) Done() {
	h.done <- struct{}{}
}

type Handler struct {
	ch chan HttpContext
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch := make(chan struct{})
	defer close(ch)

	h.ch <- HttpContext{Writer: w, Request: r, done: ch}
	<-ch
}

type Http struct {
	*http.Server
	ch chan HttpContext
}

func NewHttp(s *http.Server, pattern string) chord.Trigger[HttpContext] {
	ch := make(chan HttpContext)

	h := Handler{ch}

	var mux *http.ServeMux

	if existingMux, ok := s.Handler.(*http.ServeMux); ok {
		mux = existingMux
	} else {
		mux = http.NewServeMux()
	}

	mux.Handle(pattern, h)
	s.Handler = mux

	return Http{s, ch}
}

func (ht Http) Stage(ctx context.Context) chord.Stage[HttpContext] {
	return func() <-chan conduit.Result[HttpContext] {
		ch := make(chan conduit.Result[HttpContext])
		go func() {
			defer ht.Close()
			defer close(ch)
			defer close(ht.ch)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					ch <- conduit.Ok(ctx, <-ht.ch)
				}
			}
		}()
		return ch
	}
}
