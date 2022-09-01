package handler

import (
	"context"
	"net/http"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

type (
	requestContextKey struct{}
)

type responseWriterAdapter struct {
	fw fsthttp.ResponseWriter
}

func (w *responseWriterAdapter) Header() http.Header {
	return http.Header(w.fw.Header())
}

func (w *responseWriterAdapter) Write(b []byte) (int, error) {
	return w.fw.Write(b)
}

func (w *responseWriterAdapter) WriteHeader(status int) {
	w.fw.WriteHeader(status)
}

func Adapt(h http.Handler) fsthttp.Handler {
	return fsthttp.HandlerFunc(func(ctx context.Context, fw fsthttp.ResponseWriter, freq *fsthttp.Request) {
		ctx = context.WithValue(ctx, requestContextKey{}, freq)

		w := &responseWriterAdapter{fw: fw}

		req, err := http.NewRequestWithContext(ctx, freq.Method, freq.URL.String(), freq.Body)
		if err != nil {
			fw.WriteHeader(fsthttp.StatusInternalServerError)
			return
		}
		req.Proto = freq.Proto
		req.ProtoMajor = freq.ProtoMajor
		req.ProtoMinor = freq.ProtoMinor
		req.Header = http.Header(freq.Header.Clone())
		req.Host = freq.Host
		req.RemoteAddr = freq.RemoteAddr

		// TODO: not sure?
		req.ContentLength = -1

		// TODO: translate some of fsthttp.TLSInfo into tls.ConnectionState

		h.ServeHTTP(w, req)
	})
}

func FastlyRequestFromContext(ctx context.Context) *fsthttp.Request {
	fstreq, _ := ctx.Value(requestContextKey{}).(*fsthttp.Request)
	return fstreq
}
