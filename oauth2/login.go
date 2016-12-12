package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/internal"
	"golang.org/x/oauth2"
)

// Errors which may occur on login.
var (
	ErrInvalidState = errors.New("oauth2: Invalid OAuth2 state parameter")
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
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		cookie, err := req.Cookie(config.Name)
		if err == nil {
			// add the cookie state to the ctx
			ctx = WithState(ctx, cookie.Value)
		} else {
			// add Cookie with a random state
			val := randomState()
			http.SetCookie(w, internal.NewCookie(config, val))
			ctx = WithState(ctx, val)
		}
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// http://stackoverflow.com/a/7722099
type Oauth2State struct {
	CSRF       string `json:"csrf" url:"csrf"`
	RedirectTo string `json:"redirectTo" url:"redirectTo"`
	Transport  string `json:"transport" url:"transport"` // hash, query
}

func (s *Oauth2State) Encode() string {
	data, _ := json.Marshal(s)
	return base64.URLEncoding.EncodeToString(data)
}

func (s *Oauth2State) Decode(state string) {
	data, _ := base64.URLEncoding.DecodeString(state)
	json.Unmarshal(data, s)
}

const (
	KeyRedirectTo = "redirectTo"
	KeyTransport  = "transport"
)

// HostedLoginHandler checks for a state cookie. If found, the state value is read
// and added to the ctx. Otherwise, a non-guessable value is added to the ctx
// and to a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a ContextHandler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func HostedLoginHandler(config gologin.CookieConfig, success http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		cookie, err := req.Cookie(config.Name)
		if err == nil {
			// add the cookie state to the ctx
			ctx = WithState(ctx, cookie.Value)
		} else {
			state := Oauth2State{
				CSRF:       randomState(),
				RedirectTo: req.URL.Query().Get(KeyRedirectTo),
				Transport:  req.URL.Query().Get(KeyTransport),
			}
			val := state.Encode()
			http.SetCookie(w, internal.NewCookie(config, val))
			ctx = WithState(ctx, val)
		}
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// LoginHandler handles OAuth2 login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler, opts ...oauth2.AuthCodeOption) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		state, err := StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		authURL := config.AuthCodeURL(state, opts...)
		http.Redirect(w, req, authURL, http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// CallbackHandler handles OAuth2 redirection URI requests by parsing the auth
// code and state, comparing with the state value from the ctx, and obtaining
// an OAuth2 Token.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		authCode, state, err := parseCallback(req)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ownerState, err := StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		if state != ownerState || state == "" {
			ctx = gologin.WithError(ctx, ErrInvalidState)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		// use the authorization code to get a Token
		token, err := config.Exchange(ctx, authCode)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithToken(ctx, token)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// Returns a base64 encoded random 32 byte string.
func randomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// parseCallback parses the "code" and "state" parameters from the http.Request
// and returns them.
func parseCallback(req *http.Request) (authCode, state string, err error) {
	err = req.ParseForm()
	if err != nil {
		return "", "", err
	}
	authCode = req.Form.Get("code")
	state = req.Form.Get("state")
	if authCode == "" || state == "" {
		return "", "", errors.New("oauth2: Request missing code or state")
	}
	return authCode, state, nil
}
