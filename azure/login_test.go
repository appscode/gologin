package azure

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/gologin/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestAzureHandler(t *testing.T) {
	// jsonData := `{"id": "54638001", "name": "Ivy Crimson"}`
	expectedUser := &User{ID: "54638001", Name: "Ivy Crimson"}
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx := oauth2Login.WithToken(context.Background(), anyToken)

	config := &oauth2.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		azureUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, azureUser)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// AzureHandler assert that:
	// - Token is read from the ctx and passed to the azure API
	// - azure User is obtained from the azure API
	// - success handler is called
	// - azure User is added to the ctx of the success handler
	provider, _ := NewProvider()
	azureHandler := azureHandler(config, provider.Verifier(), http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	azureHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestAzureHandler_MissingCtxToken(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Context missing Token", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// AzureHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	provider, _ := NewProvider()
	azureHandler := azureHandler(config, provider.Verifier(), success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	azureHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}
