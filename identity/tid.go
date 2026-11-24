package identity

import (
	"context"
	"errors"
)

const (
	TIDTokenMetadataKey      = "x-trc-tid-token"
	TIDTokenCacheKey         = "trcshtalk:su:tid"
	TIDValidatorProperty     = "validateTalkctlTID"
	TIDTokenProviderProperty = "talkctlTIDToken"
)

var (
	ErrTIDTokenExpired    = errors.New("TID token expired")
	ErrTIDAuthUnavailable = errors.New("TID authentication service unavailable")
)

type tidAuthorizationKey struct{}

// WithTIDAuthorization is for server-side authenticators, never wire metadata.
func WithTIDAuthorization(ctx context.Context) context.Context {
	return context.WithValue(ctx, tidAuthorizationKey{}, true)
}

func HasTIDAuthorization(ctx context.Context) bool {
	authorized, _ := ctx.Value(tidAuthorizationKey{}).(bool)
	return authorized
}
