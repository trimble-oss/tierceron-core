package cache

import (
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// CertValue holds a cached certificate and associated metadata.
type CertValue struct {
	CertBytes   *[]byte
	CreatedTime any
	NotAfter    *time.Time
	LastUpdate  *time.Time
	Sha256      string
}

// CertCache is a concurrent-safe cache for certificates.
type CertCache struct {
	cache *cmap.ConcurrentMap[string, *CertValue]
}

// NewCertCache creates a new, initialized CertCache.
func NewCertCache() *CertCache {
	ccmap := cmap.New[*CertValue]()
	return &CertCache{cache: &ccmap}
}

// Get retrieves a CertValue by key. Returns (nil, false) if the cache or key is absent.
func (cc *CertCache) Get(key string) (*CertValue, bool) {
	if cc == nil || cc.cache == nil {
		return nil, false
	}
	return cc.cache.Get(key)
}

// Set stores a CertValue under the given key. No-op if the cache is nil.
func (cc *CertCache) Set(key string, value *CertValue) {
	if cc == nil || cc.cache == nil {
		return
	}
	cc.cache.Set(key, value)
}

// Evict removes the entry for the given key. No-op if the cache is nil.
func (cc *CertCache) Evict(key string) {
	if cc == nil || cc.cache == nil {
		return
	}
	cc.cache.Remove(key)
}

// Items returns a snapshot of all cached entries. Returns nil if the cache is nil.
func (cc *CertCache) Items() map[string]*CertValue {
	if cc == nil || cc.cache == nil {
		return nil
	}
	return cc.cache.Items()
}

func (cc *CertCache) IsEmpty() bool {
	return cc.cache.IsEmpty() || len(cc.cache.Items()) == 0
}
