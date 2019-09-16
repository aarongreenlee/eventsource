// Package session stubs a package for our event source example which would
// provide session management and accessors to values stored in context.
package session

import (
	"context"
)

type contextKey string

const (
	contextKeyUsername contextKey = "username"
	contextKeyUserID   contextKey = "userID"
)

func Stub(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, contextKeyUsername, "Cookie Monster")
	ctx = context.WithValue(ctx, contextKeyUserID, "abc123")
	return ctx
}

func Username(ctx context.Context) string {
	return ctx.Value(contextKeyUsername).(string)
}

func UserID(ctx context.Context) string {
	return ctx.Value(contextKeyUserID).(string)
}
