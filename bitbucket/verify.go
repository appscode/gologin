package bitbucket

import (
	"net/http"

	"github.com/dghubble/sling"
)

const bitbucketAPI = "https://bitbucket.org/api/2.0/"

// User is a Bitbucket user.
type User struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Links       struct {
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
	Website          string `json:"website"`
	Email            string `json:"email"`
	IsEmailConfirmed bool   `json:"is_confirmed"`
}

type UserEmails struct {
	Values []struct {
		Email       string `json:"email"`
		IsConfirmed bool   `json:"is_confirmed"`
		IsPrimary   bool   `json:"is_primary"`
	} `json:"values"`
}

// client is a Bitbucket client for obtaining a User.
type client struct {
	sling *sling.Sling
}

// newClient returns a new Bitbucket client.
func newClient(httpClient *http.Client) *client {
	base := sling.New().Client(httpClient).Base(bitbucketAPI)
	return &client{
		sling: base,
	}
}

// CurrentUser gets the current user's profile information.
// https://confluence.atlassian.com/bitbucket/users-endpoint-423626336.html
func (c *client) CurrentUser() (*User, *http.Response, error) {
	user := new(User)
	resp, err := c.sling.New().Get("user").ReceiveSuccess(user)
	if err == nil {
		emails := new(UserEmails)
		_, err = c.sling.New().Get("user/emails").ReceiveSuccess(emails)
		for _, email := range emails.Values {
			if email.IsPrimary {
				user.Email = email.Email
				user.IsEmailConfirmed = email.IsConfirmed
				break
			}
		}
	}
	return user, resp, err
}
