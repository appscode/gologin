package azure

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextUser(t *testing.T) {
	expectedUser := &User{ID: "12", Name: "Gopher"}
	ctx := WithUser(context.Background(), expectedUser)
	user, err := UserFromContext(ctx)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
}

func TestContextUser_Error(t *testing.T) {
	user, err := UserFromContext(context.Background())
	assert.Nil(t, user)
	if assert.NotNil(t, err) {
		assert.Equal(t, "azure: Context missing Azure Active Directory User", err.Error())
	}
}
