package amazon

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dghubble/gologin/testutils"
)

// newAmazonTestServer returns a new httptest.Server which mocks the Amazon
// user endpoint and a client which proxies requests to the server. The server
// responds with the given json data. The caller must close the server.
func newAmazonTestServer(jsonData string) (*http.Client, *httptest.Server) {
	client, mux, server := testutils.TestServer()
	mux.HandleFunc("/v2.4/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, server
}
