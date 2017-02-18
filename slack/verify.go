package slack

import (
	"net/http"

	"github.com/dghubble/sling"
	"golang.org/x/oauth2"
)

const slackAPI = "https://slack.com/api/"

// User is a Slack user.
//
// Note that user ids are unique to each app.
type User struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Email    string `json:"email"`
	Image24  string `json:"image_24"`
	Image32  string `json:"image_32"`
	Image48  string `json:"image_48"`
	Image72  string `json:"image_72"`
	Image192 string `json:"image_192"`
	Image512 string `json:"image_512"`
	Team     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
}

// client is a Slack client for obtaining the current User.
type client struct {
	c     *http.Client
	sling *sling.Sling
	t     *oauth2.Token
}

func newClient(httpClient *http.Client, t *oauth2.Token) *client {
	base := sling.New().Client(httpClient).Base(slackAPI)
	return &client{
		c:     httpClient,
		sling: base,
		t:     t,
	}
}

func (c *client) Profile() (*User, *http.Response, error) {
	type Params struct {
		Token string `url:"token,omitempty"`
	}
	q := &Params{
		Token: c.t.AccessToken,
	}

	type UserIdentity struct {
		Ok   bool `json:"ok"`
		User struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Image24  string `json:"image_24"`
			Image32  string `json:"image_32"`
			Image48  string `json:"image_48"`
			Image72  string `json:"image_72"`
			Image192 string `json:"image_192"`
			Image512 string `json:"image_512"`
		} `json:"user"`
		Team struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
	}
	ui := new(UserIdentity)

	// Slack returns JSON as Content-Type text/javascript :(
	// Set Accept header to receive proper Content-Type application/json
	// so Sling will decode into the struct
	resp, err := c.sling.New().Set("Accept", "application/json").Get("users.identity").QueryStruct(q).ReceiveSuccess(ui)
	user := new(User)
	if ui.Ok {
		user.ID = ui.User.ID
		user.Name = ui.User.Name
		user.Email = ui.User.Email
		user.Image24 = ui.User.Image24
		user.Image32 = ui.User.Image32
		user.Image48 = ui.User.Image48
		user.Image72 = ui.User.Image72
		user.Image192 = ui.User.Image192
		user.Image512 = ui.User.Image512

		user.Team.ID = ui.Team.ID
		user.Team.Name = ui.Team.Name
	}
	return user, resp, err
}
