package session

import (
	"context"
	"errors"
	// "fmt"
)

type Session struct {
	UserID uint64
	Login string
}

func NewSession(userID uint64, login string) *Session {
	return &Session{
		UserID: userID,
		Login: login,
	}
}

var (
	ErrNoAuth = errors.New("no session found")
)

type sessKey string

var SessionKey sessKey = "sessionKey"

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(SessionKey).(*Session)
	if !ok || sess == nil {
		// fmt.Println(ok, sess)
		return nil, ErrNoAuth
	}
	return sess, nil
}

func ContextWithSession(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, SessionKey, sess)
}
