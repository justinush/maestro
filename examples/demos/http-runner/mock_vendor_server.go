package main

import (
	"net/http"
	"net/http/httptest"
)

// newVendorMock returns a server that mimics a small vendor liveness API (offline, in-process).
// Use srv.Client() with RegistryWithHTTP so requests hit this server, not the public internet.
func newVendorMock() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/liveness/status":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"pass"}`))
		default:
			http.NotFound(w, r)
		}
	}))
}
