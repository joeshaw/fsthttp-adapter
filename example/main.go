package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/fastly/compute-sdk-go/fsthttp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joeshaw/fsthttp-adapter/handler"
)

const backend = "ipv4"

func main() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middleware.Timeout(5 * time.Second))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	router.Get("/ip", func(w http.ResponseWriter, r *http.Request) {
		req, err := fsthttp.NewRequest("GET", "https://ipv4.joeshaw.org/ip", nil)
		if err != nil {
			w.WriteHeader(fsthttp.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		req.Header.Set("Request-ID", middleware.GetReqID(r.Context()))
		req.Header.Set("Fastly-Debug", "1")

		resp, err := req.Send(r.Context(), backend)
		if err != nil {
			w.WriteHeader(fsthttp.StatusBadGateway)
			w.Write([]byte(err.Error()))
			return
		}

		for k, v := range resp.Header {
			w.Header()[k] = v
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		fmt.Fprintf(w, "\n---\n")

		ofr := handler.FastlyRequestFromContext(r.Context())
		fmt.Fprintf(w, "%s\n", ofr.Host)
	})

	router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("oh no")
	})

	router.Get("/long", func(w http.ResponseWriter, r *http.Request) {
		rand.Seed(time.Now().UnixNano())

		ctx := r.Context()
		processTime := time.Duration(rand.Intn(10)+1) * time.Second

		select {
		case <-ctx.Done():
			return

		case <-time.After(processTime):
			// The above channel simulates some hard work.
		}

		w.Write([]byte("done"))
	})

	fsthttp.Serve(handler.Adapt(router))
}
