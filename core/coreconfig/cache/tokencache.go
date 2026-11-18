package cache

import (
	"errors"
	"strings"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/trimble-oss/tierceron-core/v2/buildopts/memonly"
	"github.com/trimble-oss/tierceron-core/v2/buildopts/memprotectopts"
)

type TokenCache struct {
	VaultAddressPtr *string                                // Vault address
	rcache          *cmap.ConcurrentMap[string, *[]string] // role name, role/secret
	cache           *cmap.ConcurrentMap[string, *string]   // tokenKey, *token
}

var protectMemory = memprotectopts.MemProtect

// NewTokenCacheEmpty creates an empty cache without attempting memory protection.
func NewTokenCacheEmpty() *TokenCache {
	ccmap := cmap.New[*string]()
	rmap := cmap.New[*[]string]()
	tc := &TokenCache{rcache: &rmap, cache: &ccmap}
	return tc
}

// NewTokenCache protects supplied nonempty values in memonly builds before publishing them.
// To create an address-only cache, pass an empty token key and a nil token.
func NewTokenCache(tokenKey string, token *string, vaptr *string) (*TokenCache, error) {
	tc := NewTokenCacheEmpty()
	if vaptr != nil && len(*vaptr) > 0 {
		if err := tc.SetVaultAddress(vaptr); err != nil {
			return nil, err
		}
	}
	if token != nil && len(*token) > 0 {
		if err := tc.AddToken(tokenKey, token); err != nil {
			return nil, err
		}
	}
	return tc, nil
}

func (tc *TokenCache) IsEmpty() bool {
	return tc.cache.IsEmpty() && tc.rcache.IsEmpty()
}

func (tc *TokenCache) SetVaultAddress(vaptr *string) error {
	if vaptr == nil || len(*vaptr) == 0 {
		return errors.New("Vault address nil or empty")
	}
	if memonly.IsMemonly() {
		if err := protectMemory(nil, vaptr); err != nil {
			return err
		}
	}
	tc.VaultAddressPtr = vaptr
	return nil
}

func (tc *TokenCache) AddRoleStr(roleKey string, role *string) error {
	if len(roleKey) == 0 {
		return errors.New("key cannot be empty")
	}
	if role == nil || len(*role) == 0 {
		return errors.New("role nil or empty")
	}
	if memonly.IsMemonly() {
		if err := protectMemory(nil, role); err != nil {
			return err
		}
	}

	roleSlice := strings.Split(*role, ":")
	return tc.AddRole(roleKey, &roleSlice)
}

func (tc *TokenCache) AddRole(roleKey string, roleSlice *[]string) error {
	if len(roleKey) == 0 {
		return errors.New("key cannot be empty")
	}
	if roleSlice == nil || len(*roleSlice) == 0 {
		return errors.New("role nil or empty")
	}
	if memonly.IsMemonly() {
		for i := range len(*roleSlice) {
			if err := protectMemory(nil, &(*roleSlice)[i]); err != nil {
				return err
			}
		}
	}
	tc.rcache.Set(roleKey, roleSlice)
	return nil
}

func (tc *TokenCache) GetRole(roleKey string) *[]string {
	return tc.GetRoleStr(&roleKey)
}

func (tc *TokenCache) GetRoleWithDefault(roleKey string, defaultRole string) *[]string {
	return tc.GetRoleStr(&roleKey)
}

func (tc *TokenCache) GetRoleStr(roleKey *string) *[]string {
	if roleKey == nil {
		return nil
	}
	if tc.rcache == nil {
		return nil
	}
	if role, ok := tc.rcache.Get(*roleKey); ok {
		return role
	} else {
		return nil
	}
}

func (tc *TokenCache) AddToken(tokenKey string, token *string) error {
	if len(tokenKey) == 0 {
		return errors.New("key cannot be empty")
	}
	if token == nil || len(*token) == 0 {
		return errors.New("token nil or empty")
	}
	if memonly.IsMemonly() {
		if err := protectMemory(nil, token); err != nil {
			return err
		}
	}
	tc.cache.Set(tokenKey, token)
	return nil
}

func (tc *TokenCache) GetToken(tokenKey string) *string {
	if tc.cache == nil {
		return nil
	}
	if token, ok := tc.cache.Get(tokenKey); ok {
		return token
	} else {
		return nil
	}
}

func (tc *TokenCache) GetTokenStr(tokenKeyPtr *string) *string {
	return tc.GetToken(*tokenKeyPtr)
}

func (tc *TokenCache) Clear() {
	tc.cache.Clear()
}

func (tc *TokenCache) RemoveToken(tokenKey string) {
	if len(tokenKey) > 0 && tc.cache != nil {
		tc.cache.Remove(tokenKey)
	}
}
