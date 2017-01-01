package amazon

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

// unexported key type prevents collisions
type key int

const (
	userKey key = iota
)

// Endpoint is Amazon's OAuth 2.0 endpoint.
var Endpoint = oauth2.Endpoint{
	AuthURL:  "https://www.amazon.com/ap/oa",
	TokenURL: "https://api.amazon.com/auth/o2/token",
}

// WithUser returns a copy of ctx that stores the Amazon User.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Amazon User from the ctx.
func UserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(userKey).(*User)
	if !ok {
		return nil, fmt.Errorf("amazon: Context missing Amazon User")
	}
	return user, nil
}
