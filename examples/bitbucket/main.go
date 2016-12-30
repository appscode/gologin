package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/bitbucket"
	"github.com/dghubble/sessions"
	"golang.org/x/oauth2"
	bitbucketOAuth2 "golang.org/x/oauth2/bitbucket"
)

const (
	sessionName    = "example-bitbucket-app"
	sessionSecret  = "example cookie signing secret"
	sessionUserKey = "bitbucketID"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

// Config configures the main ServeMux.
type Config struct {
	BitBucketClientID     string
	BitBucketClientSecret string
}

// New returns a new ServeMux with app routes.
func New(config *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.Handle("/profile", requireLogin(http.HandlerFunc(profileHandler)))
	mux.HandleFunc("/logout", logoutHandler)
	// 1. Register Login and Callback handlers
	oauth2Config := &oauth2.Config{
		ClientID:     config.BitBucketClientID,
		ClientSecret: config.BitBucketClientSecret,
		RedirectURL:  "http://localhost:8080/bitbucket/callback",
		Endpoint:     bitbucketOAuth2.Endpoint,
		Scopes:       []string{"account", "email"},
	}
	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := gologin.DebugOnlyCookieConfig
	mux.Handle("/bitbucket/login", bitbucket.CSRFHandler(stateConfig, bitbucket.LoginHandler(oauth2Config, nil)))
	mux.Handle("/bitbucket/callback", bitbucket.CSRFHandler(stateConfig, bitbucket.CallbackHandler(oauth2Config, issueSession(), nil)))
	return mux
}

// issueSession issues a cookie session after successful BitBucket login
func issueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		bitbucketUser, err := bitbucket.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 2. Implement a success handler to issue some form of session
		session := sessionStore.New(sessionName)
		session.Values[sessionUserKey] = bitbucketUser.Username
		session.Save(w)
		http.Redirect(w, req, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// welcomeHandler shows a welcome message and login button.
func welcomeHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	if isAuthenticated(req) {
		http.Redirect(w, req, "/profile", http.StatusFound)
		return
	}
	page, _ := ioutil.ReadFile("home.html")
	fmt.Fprintf(w, string(page))
}

// profileHandler shows protected user content.
func profileHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, `<p>You are logged in!</p><form action="/logout" method="post"><input type="submit" value="Logout"></form>`)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func logoutHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		sessionStore.Destroy(w, sessionName)
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

// requireLogin redirects unauthenticated users to the login route.
func requireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if !isAuthenticated(req) {
			http.Redirect(w, req, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// isAuthenticated returns true if the user has a signed session cookie.
func isAuthenticated(req *http.Request) bool {
	if _, err := sessionStore.Get(req, sessionName); err == nil {
		return true
	}
	return false
}

// main creates and starts a Server listening.
func main() {
	const address = "localhost:8080"
	// read credentials from environment variables if available
	config := &Config{
		BitBucketClientID:     os.Getenv("BITBUCKET_CLIENT_ID"),
		BitBucketClientSecret: os.Getenv("BITBUCKET_CLIENT_SECRET"),
	}
	// allow consumer credential flags to override config fields
	clientID := flag.String("client-id", "", "BitBucket Client ID")
	clientSecret := flag.String("client-secret", "", "BitBucket Client Secret")
	flag.Parse()
	if *clientID != "" {
		config.BitBucketClientID = *clientID
	}
	if *clientSecret != "" {
		config.BitBucketClientSecret = *clientSecret
	}
	if config.BitBucketClientID == "" {
		log.Fatal("Missing BitBucket Client ID")
	}
	if config.BitBucketClientSecret == "" {
		log.Fatal("Missing BitBucket Client Secret")
	}

	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, New(config))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
