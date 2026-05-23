package simulate

import (
	"net/http"
	"time"
)

// simulateHTTPClient is used by the maestro simulate CLI only (long timeout for scripted runs).
// Embedders should pass their own *http.Client to engine.RegistryWithHTTP.
func simulateHTTPClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}
