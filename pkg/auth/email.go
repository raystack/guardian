package auth

import "context"

type AuthenticatedUserEmailContextKey struct{}

func WrapEmailInCtx(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, AuthenticatedUserEmailContextKey{}, email)
}

func FetchEmailFromCtx(ctx context.Context) string {
	if email, ok := ctx.Value(AuthenticatedUserEmailContextKey{}).(string); ok {
		return email
	}
	return ""
}
