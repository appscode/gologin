package amazon

import (
	"net/http"

	"github.com/dghubble/sling"
)

const amazonAPI = "https://api.amazon.com/user/"

// User is a Amazon user.
//
// Note that user ids are unique to each app.
type User struct {
	ID       string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Location string `json:"postal_code"`
}

// client is a Amazon client for obtaining the current User.
type client struct {
	c     *http.Client
	sling *sling.Sling
}

func newClient(httpClient *http.Client) *client {
	base := sling.New().Client(httpClient).Base(amazonAPI)
	return &client{
		c:     httpClient,
		sling: base,
	}
}

func (c *client) Profile() (*User, *http.Response, error) {
	user := new(User)
	// Amazon returns JSON as Content-Type text/javascript :(
	// Set Accept header to receive proper Content-Type application/json
	// so Sling will decode into the struct
	resp, err := c.sling.New().Set("Accept", "application/json").Get("profile").ReceiveSuccess(user)
	return user, resp, err
}
