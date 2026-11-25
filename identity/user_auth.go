package identity

import (
	"context"
	"errors"
)

const (
	OIDCIDTokenMetadataKey       = "x-trc-oidc-id-token"
	OIDCIDTokenCacheKey          = "trcshtalk:su:oidc-id-token"
	OIDCIDTokenValidatorProperty = "validateTalkctlOIDCIDToken"
	OIDCIDTokenProviderProperty  = "talkctlOIDCIDToken"
)

var (
	ErrOIDCIDTokenExpired  = errors.New("OIDC ID token expired")
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

func OIDCIDTokenProviderFromProperties(properties map[string]any) (func(context.Context, bool) (string, error), bool) {
	provider, ok := properties[OIDCIDTokenProviderProperty].(func(context.Context, bool) (string, error))
	return provider, ok && provider != nil
}

func OIDCIDTokenValidatorFromProperties(properties map[string]any) (func(context.Context, string) error, bool) {
	validator, ok := properties[OIDCIDTokenValidatorProperty].(func(context.Context, string) error)
	return validator, ok && validator != nil
}
