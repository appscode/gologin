package linkedin

import (
	"net/http"

	"github.com/dghubble/sling"
)

const linkedinAPI = "https://api.linkedin.com"

// User is a Linkedin user.
//
// Note that user ids are unique to each app.
// ref: https://developer.linkedin.com/docs/fields/basic-profile
type User struct {
	ID           string `json:"id"`
	EmailAddress string `json:"emailAddress"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	PictureURL   string `json:"pictureUrl"`
}

// client is a Linkedin client for obtaining the current User.
type client struct {
	c     *http.Client
	sling *sling.Sling
}

func newClient(httpClient *http.Client) *client {
	base := sling.New().Client(httpClient).Base(linkedinAPI)
	return &client{
		c:     httpClient,
		sling: base,
	}
}

func (c *client) People() (*User, *http.Response, error) {
	// API Console: https://apigee.com/console/linkedin
	// https://api.linkedin.com/v1/people/~:(id,firstName,lastName,email-address,picture-url)?format=json
	user := new(User)
	// Linkedin returns JSON as Content-Type text/javascript :(
	// Set Accept header to receive proper Content-Type application/json
	// so Sling will decode into the struct
	resp, err := c.sling.New().Set("Accept", "application/json").Get("/v1/people/~:(id,firstName,lastName,email-address,picture-url)?format=json").ReceiveSuccess(user)
	return user, resp, err
}
