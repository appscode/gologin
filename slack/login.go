package slack

import (
	"errors"
	"net/http"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"golang.org/x/oauth2"
	"time"
)

// Slack login errors
var (
	ErrUnableToGetSlackUser = errors.New("slack: unable to get Slack User")
)

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

// LoginHandler handles Slack login requests by reading the state value
// from the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Slack redirection URI requests and adds the
// Slack access token and User to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure
// handler.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = slackHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// slackHandler is a http.Handler that gets the OAuth2 Token from the ctx
// to get the corresponding Slack User. If successful, the user is added to
// the ctx and the success handler is called. Otherwise, the failure handler
// is called.
func slackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
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
		httpClient := &http.Client{Timeout: time.Second * 5}
		slackService := newClient(httpClient, token)
		user, resp, err := slackService.Profile()
		err = validateResponse(user, resp, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithUser(ctx, user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateResponse returns an error if the given Slack User, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetSlackUser
	}
	if user == nil || user.ID == "" {
		return ErrUnableToGetSlackUser
	}
	return nil
}
