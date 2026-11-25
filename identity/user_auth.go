package identity

import (
	"context"
	"errors"
)

const (
	UserTokenMetadataKey       = "x-trc-user-token"
	UserTokenCacheKey          = "trcshtalk:su:user-token"
	UserTokenValidatorProperty = "validateTalkctlUserToken"
	UserTokenProviderProperty  = "talkctlUserToken"
)

var (
	ErrUserTokenExpired    = errors.New("user token expired")
	ErrUserAuthUnavailable = errors.New("user authentication service unavailable")
)

type elevatedUserAuthorizationKey struct{}

// WithElevatedUserAuthorization is for server-side authenticators that have
// verified elevated access, never for unverified wire metadata.
func WithElevatedUserAuthorization(ctx context.Context) context.Context {
	return context.WithValue(ctx, elevatedUserAuthorizationKey{}, true)
}

func HasElevatedUserAuthorization(ctx context.Context) bool {
	authorized, _ := ctx.Value(elevatedUserAuthorizationKey{}).(bool)
	return authorized
}

func UserTokenProviderFromProperties(properties map[string]any) (func(context.Context, bool) (string, error), bool) {
	provider, ok := properties[UserTokenProviderProperty].(func(context.Context, bool) (string, error))
	return provider, ok && provider != nil
}

func UserTokenValidatorFromProperties(properties map[string]any) (func(context.Context, string) error, bool) {
	validator, ok := properties[UserTokenValidatorProperty].(func(context.Context, string) error)
	return validator, ok && validator != nil
}
