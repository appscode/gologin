package azure

import (
	"context"
	"errors"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"golang.org/x/oauth2"
)

// Azure login errors
var (
	ErrUnableToGetAzureUser = errors.New("azure: unable to get Azure Active Directory User")
)

// ref: https://docs.microsoft.com/en-us/azure/active-directory/active-directory-v2-flows#web-apps
// ref: https://docs.microsoft.com/en-us/azure/active-directory/active-directory-v2-tokens
func NewProvider() (*oidc.Provider, error) {
	return oidc.NewProvider(context.Background(), "https://login.microsoftonline.com/9188040d-6c67-4c5b-b112-36a304b66dad/v2.0")
}

// CSRFHandler checks for a state cookie. If found, the state value is read
// and added to the ctx. Otherwise, a non-guessable value is added to the ctx
// and to a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a http.Handler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func CSRFHandler(config gologin.CookieConfig, success http.Handler) http.Handler {
	return oauth2Login.CSRFHandler(config, success)
}

// LoginHandler handles Azure login requests by reading the state value
// from the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Azure redirection URI requests and adds the
// Azure access token and User to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure
// handler.
func CallbackHandler(config *oauth2.Config, verifier *oidc.IDTokenVerifier, success, failure http.Handler) http.Handler {
	success = azureHandler(config, verifier, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// azureHandler is a http.Handler that gets the OAuth2 Token from the ctx
// to get the corresponding Azure User. If successful, the user is added to
// the ctx and the success handler is called. Otherwise, the failure handler
// is called.
func azureHandler(config *oauth2.Config, verifier *oidc.IDTokenVerifier, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := oauth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			ctx = gologin.WithError(ctx, errors.New("No id_token field in oauth2 token."))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			ctx = gologin.WithError(ctx, errors.New("Failed to verify ID Token: %v"+err.Error()))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// Extract custom claims
		var user User
		if err := idToken.Claims(&user); err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithUser(ctx, &user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
