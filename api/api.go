package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

var (
	// Global cache for API callers
	callerCache      = make(map[string]*APICaller)
	callerCacheMutex sync.RWMutex
)

// EndpointType represents the type of API endpoint
type EndpointType string

const (
	// EndpointTypeREST represents a REST API endpoint
	EndpointTypeREST EndpointType = "rest"
	// EndpointTypeGRPC represents a gRPC API endpoint
	EndpointTypeGRPC EndpointType = "grpc"
	// EndpointTypeSOAP represents a SOAP API endpoint
	EndpointTypeSOAP EndpointType = "soap"
)

// Endpoint represents an API endpoint configuration
type Endpoint struct {
	// FriendlyName is a human-readable name for the endpoint
	FriendlyName string
	// URL is the endpoint URL/address
	URL string
	// Type is the type of endpoint (REST, gRPC, SOAP)
	Type EndpointType
	// Timeout is the timeout duration for API calls
	// If set to 0 or unset, defaults to 30 seconds
	// If set to -1, no timeout is applied (uses context.Background())
	Timeout time.Duration
	// WSDLUrl is the WSDL URL for SOAP endpoints (optional)
	// Used for service discovery and documentation
	WSDLUrl string
	// MethodName is the full method name for gRPC endpoints (e.g., "/package.Service/Method")
	// Required for gRPC calls
	MethodName string
	// MaxRetries is the maximum number of retry attempts for timeout errors
	// If set to 0 or unset, no retries are attempted
	// Retries use exponential backoff: 1s, 2s, 4s, 8s, etc.
	MaxRetries int
}

// Call makes a generic API call to this endpoint
// Parameters:
//   - ctx: Context for the request (optional, will create a default 30s timeout if nil)
//   - params: Map of parameters that will be mapped based on endpoint type
//     Common parameters across all types:
//   - "method" (string): HTTP method for REST, operation name for SOAP/gRPC
//   - "body" (any): Request body
//     REST-specific parameters:
//   - "headers" (map[string]string): HTTP headers
//     SOAP-specific parameters:
//   - "soapAction" (string): SOAP action header
//   - "headers" (map[string]string): Additional HTTP headers
//   - config: Optional API caller configuration (TLS, certificates, etc.)
//
// Returns a map with the following keys:
//   - "statusCode" (int): HTTP status code (REST/SOAP only)
//   - "body" (any): Response body (parsed as JSON if possible, otherwise raw bytes)
//   - "headers" (map[string][]string): Response headers (REST/SOAP only)
//   - "error" (string): Error message if the call failed
func (e *Endpoint) Call(params map[string]any, config *APICallerConfig) (map[string]any, error) {
	// Create and manage context internally based on Timeout setting
	var ctx context.Context
	var cancel context.CancelFunc

	if e.Timeout == -1 {
		// No timeout - use background context
		ctx = context.Background()
	} else if e.Timeout > 0 {
		// Use specified timeout
		ctx, cancel = context.WithTimeout(context.Background(), e.Timeout)
		defer cancel()
	} else {
		// Default to 30 seconds
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	if params == nil {
		params = make(map[string]any)
	}

	if config == nil {
		config = &APICallerConfig{}
	}

	// Get or create cached API caller for this endpoint
	caller, err := GetOrCreateAPICaller(*e, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get API caller: %w", err)
	}

	// Extract common parameters
	method, _ := params["method"].(string)
	body := params["body"]

	// For gRPC, use MethodName from endpoint if method not provided in params
	if e.Type == EndpointTypeGRPC && method == "" {
		method = e.MethodName
	}

	// Build call options
	callOptions := &CallOptions{
		Context: ctx,
		Method:  method,
		Body:    body,
	}

	// Extract headers if present
	if headersParam, ok := params["headers"]; ok {
		if headers, ok := headersParam.(map[string]string); ok {
			callOptions.Headers = headers
		}
	}

	// Handle SOAP-specific parameters
	if e.Type == EndpointTypeSOAP {
		if callOptions.Headers == nil {
			callOptions.Headers = make(map[string]string)
		}
		// Add SOAPAction header if specified
		if soapAction, ok := params["soapAction"].(string); ok {
			callOptions.Headers["SOAPAction"] = fmt.Sprintf(`"%s"`, soapAction)
		}
	}

	// Make the call with retry logic for timeouts
	var response *Response
	var callErr error
	maxRetries := e.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Recreate context for each retry attempt
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s, etc.
			backoffDuration := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoffDuration)

			// Recreate context with timeout for retry
			if e.Timeout == -1 {
				ctx = context.Background()
			} else if e.Timeout > 0 {
				ctx, cancel = context.WithTimeout(context.Background(), e.Timeout)
				defer cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
			}
			callOptions.Context = ctx
		}

		response, callErr = caller.Call(callOptions)

		// Check if error is a timeout error
		if callErr != nil && attempt < maxRetries {
			// Check for timeout errors
			if errors.Is(callErr, context.DeadlineExceeded) ||
				(response != nil && response.Error != nil && errors.Is(response.Error, context.DeadlineExceeded)) {
				// Timeout occurred, will retry
				continue
			}
		}

		// Success or non-timeout error, break out of retry loop
		break
	}

	// Build response map
	result := make(map[string]any)

	if response != nil {
		result["statusCode"] = response.StatusCode
		result["headers"] = response.Headers

		// Handle response body based on endpoint type
		if response.Body != nil {
			// Check if body is already parsed (e.g., from gRPC)
			if bodyMap, ok := response.Body.(map[string]any); ok {
				result["body"] = bodyMap
			} else if bodyBytes, ok := response.Body.([]byte); ok && len(bodyBytes) > 0 {
				if e.Type == EndpointTypeSOAP {
					// Parse SOAP response into key-value pairs
					parsed, err := parseSOAPResponse(bodyBytes)
					if err == nil && parsed != nil {
						result["body"] = parsed
					} else {
						// Fallback to string if parsing fails
						result["body"] = string(bodyBytes)
					}
				} else {
					// For REST, try to parse as JSON
					var jsonBody any
					if jsonErr := json.Unmarshal(bodyBytes, &jsonBody); jsonErr == nil {
						result["body"] = jsonBody
					} else {
						// Return as string if not valid JSON
						result["body"] = string(bodyBytes)
					}
				}
			} else {
				// For other types, return as-is
				result["body"] = response.Body
			}
		} else {
			result["body"] = nil
		}

		if response.Error != nil {
			result["error"] = response.Error.Error()
		}
	}

	if callErr != nil {
		result["error"] = callErr.Error()
		return result, callErr
	}

	return result, nil
}

// CallOptions represents options for making an API call
type CallOptions struct {
	// Context for the request (managed internally by Call method)
	Context context.Context
	// Method is the HTTP method for REST (GET, POST, etc.) or operation name for SOAP/gRPC
	Method string
	// Headers for REST/SOAP requests
	Headers map[string]string
	// Body is the request body
	Body any
	// Timeout for the request (optional, will use context deadline if set)
	Timeout any
}

// Response represents an API response
type Response struct {
	// StatusCode is the HTTP status code for REST/SOAP
	StatusCode int
	// Body is the response body (can be []byte, map[string]any, or other types)
	Body any
	// Headers are the response headers for REST/SOAP
	Headers map[string][]string
	// Error if the call failed
	Error error
}

// Client is the interface for making API calls
type Client interface {
	// Call makes an API call to the configured endpoint
	Call(options *CallOptions) (*Response, error)
	// Close closes the client and releases resources
	Close() error
}

// APICallerConfig represents configuration for the API caller
type APICallerConfig struct {
	// InsecureSkipVerify skips TLS certificate verification (use with caution)
	InsecureSkipVerify bool
	// TLSCertPath is the path to TLS certificate file
	TLSCertPath string
	// TLSKeyPath is the path to TLS key file
	TLSKeyPath string
	// CACertPath is the path to CA certificate file
	CACertPath string
}

// APICaller provides a generic interface for calling different types of APIs
type APICaller struct {
	endpoint Endpoint
	client   Client
	config   *APICallerConfig
	cacheKey string
}

// generateCacheKey creates a unique key for caching based on endpoint and config
func generateCacheKey(endpoint Endpoint, config *APICallerConfig) string {
	key := fmt.Sprintf("%s|%s|%s|%t|%s|%s|%s",
		endpoint.FriendlyName,
		endpoint.URL,
		endpoint.Type,
		config.InsecureSkipVerify,
		config.TLSCertPath,
		config.TLSKeyPath,
		config.CACertPath,
	)
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// GetOrCreateAPICaller returns a cached caller or creates a new one
func GetOrCreateAPICaller(endpoint Endpoint, config *APICallerConfig) (*APICaller, error) {
	if config == nil {
		config = &APICallerConfig{}
	}

	cacheKey := generateCacheKey(endpoint, config)

	// Try to get from cache first
	callerCacheMutex.RLock()
	if caller, exists := callerCache[cacheKey]; exists {
		callerCacheMutex.RUnlock()
		return caller, nil
	}
	callerCacheMutex.RUnlock()

	// Not in cache, create new caller
	callerCacheMutex.Lock()
	defer callerCacheMutex.Unlock()

	// Double-check in case another goroutine created it
	if caller, exists := callerCache[cacheKey]; exists {
		return caller, nil
	}

	// Create new caller
	caller, err := newAPICallerInternal(endpoint, config, cacheKey)
	if err != nil {
		return nil, err
	}

	// Store in cache
	callerCache[cacheKey] = caller

	return caller, nil
}

// NewAPICaller creates a new API caller for the specified endpoint (uses cache)
func NewAPICaller(endpoint Endpoint, config *APICallerConfig) (*APICaller, error) {
	return GetOrCreateAPICaller(endpoint, config)
}

// newAPICallerInternal creates a new API caller without caching (internal use)
func newAPICallerInternal(endpoint Endpoint, config *APICallerConfig, cacheKey string) (*APICaller, error) {
	if endpoint.FriendlyName == "" {
		return nil, errors.New("endpoint friendly name is required")
	}
	if endpoint.URL == "" {
		return nil, errors.New("endpoint URL is required")
	}

	if config == nil {
		config = &APICallerConfig{}
	}

	caller := &APICaller{
		endpoint: endpoint,
		config:   config,
		cacheKey: cacheKey,
	}

	// Create appropriate client based on endpoint type
	var err error
	switch endpoint.Type {
	case EndpointTypeREST:
		caller.client, err = NewRESTClient(endpoint, config)
	case EndpointTypeGRPC:
		caller.client, err = NewGRPCClient(endpoint, config)
	case EndpointTypeSOAP:
		caller.client, err = NewSOAPClient(endpoint, config)
	default:
		return nil, fmt.Errorf("unsupported endpoint type: %s", endpoint.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create client for endpoint '%s': %w", endpoint.FriendlyName, err)
	}

	return caller, nil
}

// ClearCallerCache clears the entire caller cache
func ClearCallerCache() {
	callerCacheMutex.Lock()
	defer callerCacheMutex.Unlock()

	// Close all cached callers
	for _, caller := range callerCache {
		if caller.client != nil {
			caller.client.Close()
		}
	}

	callerCache = make(map[string]*APICaller)
}

// RemoveCallerFromCache removes a specific caller from the cache
func RemoveCallerFromCache(endpoint Endpoint, config *APICallerConfig) {
	if config == nil {
		config = &APICallerConfig{}
	}

	cacheKey := generateCacheKey(endpoint, config)

	callerCacheMutex.Lock()
	defer callerCacheMutex.Unlock()

	if caller, exists := callerCache[cacheKey]; exists {
		if caller.client != nil {
			caller.client.Close()
		}
		delete(callerCache, cacheKey)
	}
}

// Call makes an API call to the configured endpoint
// The function creates and manages its own context with a default 30s timeout
func (ac *APICaller) Call(options *CallOptions) (*Response, error) {
	if ac.client == nil {
		return nil, errors.New("client not initialized")
	}

	if options == nil {
		return nil, errors.New("call options are required")
	}

	// Create and manage context internally
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	options.Context = ctx

	return ac.client.Call(options)
}

// GetEndpoint returns the configured endpoint
func (ac *APICaller) GetEndpoint() Endpoint {
	return ac.endpoint
}

// Close closes the API caller and releases resources
// Note: Callers are cached globally, so this should rarely be called directly.
// Use RemoveCallerFromCache or ClearCallerCache instead for cached callers.
func (ac *APICaller) Close() error {
	if ac.client != nil {
		return ac.client.Close()
	}
	return nil
}

// GetCacheKey returns the cache key for this caller
func (ac *APICaller) GetCacheKey() string {
	return ac.cacheKey
}

// parseSOAPResponse extracts key-value pairs from SOAP response XML
func parseSOAPResponse(xmlData []byte) (map[string]any, error) {
	// Define minimal SOAP envelope structure for parsing
	type SOAPEnvelope struct {
		XMLName xml.Name `xml:"Envelope"`
		Body    struct {
			Inner []byte `xml:",innerxml"`
		} `xml:"Body"`
	}

	// Parse SOAP envelope
	var envelope SOAPEnvelope
	if err := xml.Unmarshal(xmlData, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse SOAP envelope: %w", err)
	}

	// Parse the inner XML of the body
	if len(envelope.Body.Inner) == 0 {
		return map[string]any{}, nil
	}

	// Parse the inner body content into a generic map
	result, err := parseXMLToMap(envelope.Body.Inner)
	if err != nil {
		// If parsing fails, return raw string
		return map[string]any{
			"raw": string(envelope.Body.Inner),
		}, nil
	}

	return result, nil
}

// parseXMLToMap converts XML to a map[string]any
func parseXMLToMap(data []byte) (map[string]any, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	result := make(map[string]any)
	var current string
	var textBuf bytes.Buffer

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			current = t.Name.Local
			textBuf.Reset()
		case xml.CharData:
			textBuf.Write(t)
		case xml.EndElement:
			if current != "" && textBuf.Len() > 0 {
				text := strings.TrimSpace(textBuf.String())
				if text != "" {
					result[current] = text
				}
			}
			current = ""
			textBuf.Reset()
		}
	}

	return result, nil
}
