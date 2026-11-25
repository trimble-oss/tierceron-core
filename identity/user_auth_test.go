package identity

import (
	"context"
	"testing"
)

func TestUserTokenProperties(t *testing.T) {
	properties := map[string]any{
		UserTokenProviderProperty:  func(context.Context, bool) (string, error) { return "user-token", nil },
		UserTokenValidatorProperty: func(context.Context, string) error { return nil },
	}
	provider, ok := UserTokenProviderFromProperties(properties)
	if !ok {
		t.Fatal("user token provider not found")
	}
	if token, _ := provider(context.Background(), false); token != "user-token" {
		t.Fatal("wrong user token provider")
	}
	validator, ok := UserTokenValidatorFromProperties(properties)
	if !ok || validator(context.Background(), "") != nil {
		t.Fatal("user token validator not found")
	}
	for _, invalid := range []any{nil, "invalid"} {
		properties[UserTokenProviderProperty] = invalid
		properties[UserTokenValidatorProperty] = invalid
		if _, ok := UserTokenProviderFromProperties(properties); ok {
			t.Fatal("invalid provider accepted")
		}
		if _, ok := UserTokenValidatorFromProperties(properties); ok {
			t.Fatal("invalid validator accepted")
		}
	}
}
