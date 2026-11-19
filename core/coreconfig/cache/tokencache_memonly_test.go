//go:build memonly

package cache

import (
	"errors"
	"log"
	"testing"
)

func TestProtectionFailureDoesNotPublishCredentials(t *testing.T) {
	original := protectMemory
	t.Cleanup(func() { protectMemory = original })
	want := errors.New("memory lock denied")
	protectMemory = func(*log.Logger, *string) error { return want }

	tc := NewTokenCacheEmpty()
	value := "test data"
	role := []string{"role", "secret"}
	for name, operation := range map[string]func() error{
		"address": func() error { return tc.SetVaultAddress(&value) },
		"token":   func() error { return tc.AddToken("token", &value) },
		"role":    func() error { return tc.AddRole("role", &role) },
		"roleStr": func() error { return tc.AddRoleStr("roleStr", &value) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := operation(); !errors.Is(err, want) {
				t.Fatalf("got %v; want %v", err, want)
			}
		})
	}
	if tc.VaultAddressPtr != nil || !tc.IsEmpty() {
		t.Fatal("unlocked data was published")
	}
}

func TestNewTokenCacheRejectsTokenProtectionFailure(t *testing.T) {
	original := protectMemory
	t.Cleanup(func() { protectMemory = original })
	want := errors.New("memory lock denied")
	token, address := "test token", "test address"
	protectMemory = func(_ *log.Logger, value *string) error {
		if *value == token {
			return want
		}
		return nil
	}
	tc, err := NewTokenCache("token", &token, &address)
	if tc != nil || !errors.Is(err, want) {
		t.Fatalf("got cache %p, error %v; want nil cache and lock error", tc, err)
	}
}

func TestTokenProtectionFailureDoesNotReplaceExistingToken(t *testing.T) {
	original := protectMemory
	t.Cleanup(func() { protectMemory = original })
	protectMemory = func(*log.Logger, *string) error { return nil }
	tc := NewTokenCacheEmpty()
	existing := "existing token"
	if err := tc.AddToken("token", &existing); err != nil {
		t.Fatal(err)
	}
	want := errors.New("memory lock denied")
	protectMemory = func(*log.Logger, *string) error { return want }
	replacement := "replacement token"
	if err := tc.AddToken("token", &replacement); !errors.Is(err, want) {
		t.Fatalf("got %v; want %v", err, want)
	}
	if token := tc.GetToken("token"); token == nil || *token != existing {
		t.Fatal("unprotected token replaced existing credentials")
	}
}

func TestRoleProtectionFailureDoesNotReplaceExistingRole(t *testing.T) {
	original := protectMemory
	t.Cleanup(func() { protectMemory = original })
	protectMemory = func(*log.Logger, *string) error { return nil }
	tc := NewTokenCacheEmpty()
	existing := []string{"existing role", "existing secret"}
	if err := tc.AddRole("role", &existing); err != nil {
		t.Fatal(err)
	}
	want := errors.New("memory lock denied")
	calls := 0
	protectMemory = func(*log.Logger, *string) error {
		calls++
		if calls == 2 {
			return want
		}
		return nil
	}
	replacement := []string{"replacement role", "replacement secret"}
	if err := tc.AddRole("role", &replacement); !errors.Is(err, want) {
		t.Fatalf("got %v; want %v", err, want)
	}
	if tc.GetRole("role") != &existing {
		t.Fatal("partially protected role replaced existing credentials")
	}
}
