package app

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

type helper struct {
	authenticatedUserHeaderKey string
}

func NewHelper(authenticatedUserHeaderKey string) *helper {
	return &helper{authenticatedUserHeaderKey}
}

func (h helper) GetAuthenticatedUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("unable to retrieve metadata from context")
	}

	users := md.Get(h.authenticatedUserHeaderKey)
	if len(users) == 0 {
		return "", errors.New("user email not found")
	}

	return users[0], nil
}
