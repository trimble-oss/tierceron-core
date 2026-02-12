package api

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RESTClient implements the Client interface for REST APIs
type RESTClient struct {
	endpoint   Endpoint
	httpClient *http.Client
	config     *APICallerConfig
}

// NewRESTClient creates a new REST client
func NewRESTClient(endpoint Endpoint, config *APICallerConfig) (*RESTClient, error) {
	client := &RESTClient{
		endpoint: endpoint,
		config:   config,
	}

	// Create HTTP client with TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	// Load CA certificate if provided
	if len(config.CACertData) > 0 {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(config.CACertData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate if provided
	if len(config.TLSCertData) > 0 && len(config.TLSKeyData) > 0 {
		cert, err := tls.X509KeyPair(config.TLSCertData, config.TLSKeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	client.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		// No timeout here - we use context timeouts per-request
	}

	return client, nil
}

// Call makes a REST API call
func (rc *RESTClient) Call(options *CallOptions) (*Response, error) {
	if options.Method == "" {
		options.Method = "GET"
	}

	// Prepare request body
	var bodyReader io.Reader
	if options.Body != nil {
		if rc.endpoint.Type == EndpointTypeFormURLEncoded {
			formReader, err := encodeFormBody(options.Body)
			if err != nil {
				return nil, err
			}
			bodyReader = formReader
		} else {
			switch v := options.Body.(type) {
			case []byte:
				bodyReader = bytes.NewReader(v)
			case string:
				bodyReader = bytes.NewReader([]byte(v))
			case io.Reader:
				bodyReader = v
			default:
				// Try to marshal as JSON
				jsonBody, err := json.Marshal(options.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal request body: %w", err)
				}
				bodyReader = bytes.NewReader(jsonBody)
			}
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(options.Context, options.Method, rc.endpoint.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if options.Headers != nil {
		for key, value := range options.Headers {
			req.Header.Set(key, value)
		}
	}

	// Set default Content-Type if not specified and body is present
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		if rc.endpoint.Type == EndpointTypeFormURLEncoded {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Make the request
	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return &Response{
			Error: fmt.Errorf("request failed: %w", err),
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Error:      fmt.Errorf("failed to read response body: %w", err),
		}, err
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		response.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return response, response.Error
}

func encodeFormBody(body any) (io.Reader, error) {
	switch v := body.(type) {
	case url.Values:
		return strings.NewReader(v.Encode()), nil
	case map[string]string:
		values := url.Values{}
		for key, value := range v {
			values.Set(key, value)
		}
		return strings.NewReader(values.Encode()), nil
	case map[string][]string:
		values := url.Values{}
		for key, value := range v {
			values[key] = value
		}
		return strings.NewReader(values.Encode()), nil
	case map[string]any:
		values := url.Values{}
		for key, value := range v {
			values.Set(key, fmt.Sprint(value))
		}
		return strings.NewReader(values.Encode()), nil
	case []byte:
		return bytes.NewReader(v), nil
	case string:
		return strings.NewReader(v), nil
	case io.Reader:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported form body type: %T", body)
	}
}

// Close closes the REST client
func (rc *RESTClient) Close() error {
	// HTTP client doesn't need explicit closing
	if rc.httpClient != nil {
		rc.httpClient.CloseIdleConnections()
	}
	return nil
}
