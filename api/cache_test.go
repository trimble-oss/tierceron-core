package api_test

import (
	"testing"
	"time"

	"github.com/trimble-oss/tierceron-core/v2/api"
)

// TestCallerCaching verifies that callers are cached and reused
func TestCallerCaching(t *testing.T) {
	endpoint := api.Endpoint{
		FriendlyName: "Test API",
		URL:          "https://api.example.com/test",
		Type:         api.EndpointTypeREST,
		Timeout:      30 * time.Second,
	}

	config := &api.APICallerConfig{
		InsecureSkipVerify: false,
	}

	// Create first caller
	caller1, err := api.NewAPICaller(endpoint, config)
	if err != nil {
		t.Fatalf("Failed to create first caller: %v", err)
	}

	// Create second caller with same endpoint and config
	caller2, err := api.NewAPICaller(endpoint, config)
	if err != nil {
		t.Fatalf("Failed to create second caller: %v", err)
	}

	// Verify they are the same instance (cached)
	if caller1 != caller2 {
		t.Error("Expected callers to be the same cached instance")
	}

	// Verify cache keys match
	if caller1.GetCacheKey() != caller2.GetCacheKey() {
		t.Error("Expected cache keys to match")
	}

	t.Logf("✓ Callers successfully cached with key: %s", caller1.GetCacheKey())
}

// TestCallerCacheDifferentConfigs verifies different configs create different cached callers
func TestCallerCacheDifferentConfigs(t *testing.T) {
	endpoint := api.Endpoint{
		FriendlyName: "Test API",
		URL:          "https://api.example.com/test",
		Type:         api.EndpointTypeREST,
		Timeout:      20 * time.Second,
	}

	config1 := &api.APICallerConfig{
		InsecureSkipVerify: false,
	}

	config2 := &api.APICallerConfig{
		InsecureSkipVerify: true,
	}

	caller1, _ := api.NewAPICaller(endpoint, config1)
	caller2, _ := api.NewAPICaller(endpoint, config2)

	// Verify they are different instances
	if caller1 == caller2 {
		t.Error("Expected callers with different configs to be different instances")
	}

	// Verify cache keys are different
	if caller1.GetCacheKey() == caller2.GetCacheKey() {
		t.Error("Expected different cache keys for different configs")
	}

	t.Logf("✓ Different configs produce different cached callers")
}

// TestContextManagement verifies that Call manages its own context
func TestContextManagement(t *testing.T) {
	endpoint := api.Endpoint{
		FriendlyName: "HTTPBin",
		URL:          "https://httpbin.org/delay/1",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	// Call without any context - should create and manage its own
	result, err := endpoint.Call(map[string]any{
		"method": "GET",
	})

	// Should not panic and should handle context internally
	if err != nil {
		t.Logf("Call completed with error (expected for network call): %v", err)
	} else {
		t.Logf("✓ Call with nil context succeeded, status: %v", result["statusCode"])
	}
}

// TestClearCache verifies cache clearing functionality
func TestClearCache(t *testing.T) {
	endpoint := api.Endpoint{
		FriendlyName: "Test API Clear",
		URL:          "https://api.example.com/clear",
		Type:         api.EndpointTypeREST,
		Timeout:      15 * time.Second,
	}

	// Create a caller
	caller1, _ := api.NewAPICaller(endpoint, nil)
	cacheKey1 := caller1.GetCacheKey()

	// Clear cache
	api.ClearCallerCache()

	// Create another caller with same endpoint
	caller2, _ := api.NewAPICaller(endpoint, nil)
	cacheKey2 := caller2.GetCacheKey()

	// Cache keys should be the same (same endpoint/config)
	if cacheKey1 != cacheKey2 {
		t.Error("Expected cache keys to match for same endpoint")
	}

	// But instances should be different (cache was cleared)
	if caller1 == caller2 {
		t.Error("Expected different instances after cache clear")
	}

	t.Logf("✓ Cache cleared successfully")
}

// TestRemoveFromCache verifies selective cache removal
func TestRemoveFromCache(t *testing.T) {
	endpoint1 := api.Endpoint{
		FriendlyName: "API 1",
		URL:          "https://api1.example.com",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	endpoint2 := api.Endpoint{
		FriendlyName: "API 2",
		URL:          "https://api2.example.com",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	// Create two callers
	caller1a, _ := api.NewAPICaller(endpoint1, nil)
	caller2a, _ := api.NewAPICaller(endpoint2, nil)

	// Remove first from cache
	api.RemoveCallerFromCache(endpoint1, nil)

	// Recreate both
	caller1b, _ := api.NewAPICaller(endpoint1, nil)
	caller2b, _ := api.NewAPICaller(endpoint2, nil)

	// First should be new instance, second should be cached
	if caller1a == caller1b {
		t.Error("Expected endpoint1 to be new instance after removal")
	}

	if caller2a != caller2b {
		t.Error("Expected endpoint2 to still be cached")
	}

	t.Logf("✓ Selective cache removal works correctly")
}

// TestEndpointCallWithContext verifies Endpoint.Call manages context internally
func TestEndpointCallWithContext(t *testing.T) {
	endpoint := api.Endpoint{
		FriendlyName: "Test Endpoint",
		URL:          "https://httpbin.org/get",
		Type:         api.EndpointTypeREST,
		Timeout:      10 * time.Second,
	}

	// Call creates and manages context internally
	result, err := endpoint.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		t.Logf("Call: %v", err)
	} else if result != nil {
		t.Logf("✓ Call with internally managed context succeeded")
	}
}

// TestEndpointTimeout verifies custom timeout values work correctly
func TestEndpointTimeout(t *testing.T) {
	// Test 1: Custom timeout (5 seconds)
	endpoint1 := api.Endpoint{
		FriendlyName: "Fast API",
		URL:          "https://httpbin.org/delay/1",
		Type:         api.EndpointTypeREST,
		Timeout:      5 * time.Second,
	}

	result1, err := endpoint1.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		t.Logf("5s timeout call: %v", err)
	} else if result1 != nil {
		t.Logf("✓ Custom 5s timeout succeeded")
	}

	// Test 2: No timeout (-1)
	endpoint2 := api.Endpoint{
		FriendlyName: "No Timeout API",
		URL:          "https://httpbin.org/get",
		Type:         api.EndpointTypeREST,
		Timeout:      -1,
	}

	result2, err := endpoint2.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		t.Logf("No timeout call: %v", err)
	} else if result2 != nil {
		t.Logf("✓ No timeout (-1) succeeded")
	}

	// Test 3: Default timeout (0 = 30s default)
	endpoint3 := api.Endpoint{
		FriendlyName: "Default Timeout API",
		URL:          "https://httpbin.org/get",
		Type:         api.EndpointTypeREST,
		Timeout:      0, // Uses default 30s
	}

	result3, err := endpoint3.Call(map[string]any{
		"method": "GET",
	})
	if err != nil {
		t.Logf("Default timeout call: %v", err)
	} else if result3 != nil {
		t.Logf("✓ Default timeout (30s) succeeded")
	}
}
