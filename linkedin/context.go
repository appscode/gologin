package linkedin

import (
	"context"
	"fmt"
)

// unexported key type prevents collisions
type key int

const (
	userKey key = iota
)

// WithUser returns a copy of ctx that stores the Linkedin User.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Linkedin User from the ctx.
func UserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(userKey).(*User)
	if !ok {
		return nil, fmt.Errorf("linkedin: Context missing Linkedin User")
	}
	return user, nil
}
