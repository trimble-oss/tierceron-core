package identity

import (
	"context"
	"testing"
)

func TestOIDCIDTokenProperties(t *testing.T) {
	properties := map[string]any{
		"talkctlOIDCIDToken":         func(context.Context, bool) (string, error) { return "oidc-id-token", nil },
		"validateTalkctlOIDCIDToken": func(context.Context, string) error { return nil },
	}
	provider, ok := OIDCIDTokenProviderFromProperties(properties)
	if !ok {
		t.Fatal("OIDC ID token provider not found")
	}
	if token, _ := provider(context.Background(), false); token != "oidc-id-token" {
		t.Fatal("wrong OIDC ID token provider")
	}
	validator, ok := OIDCIDTokenValidatorFromProperties(properties)
	if !ok || validator(context.Background(), "") != nil {
		t.Fatal("OIDC ID token validator not found")
	}
	for _, invalid := range []any{nil, "invalid"} {
		properties[OIDCIDTokenProviderProperty] = invalid
		properties[OIDCIDTokenValidatorProperty] = invalid
		if _, ok := OIDCIDTokenProviderFromProperties(properties); ok {
			t.Fatal("invalid provider accepted")
		}
		if _, ok := OIDCIDTokenValidatorFromProperties(properties); ok {
			t.Fatal("invalid validator accepted")
		}
	}
}
