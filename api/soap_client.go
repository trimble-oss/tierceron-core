package api

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// SOAPEnvelope represents a SOAP 1.1/1.2 envelope structure
type SOAPEnvelope struct {
	XMLName xml.Name    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  *SOAPHeader `xml:"Header,omitempty"`
	Body    SOAPBody    `xml:"Body"`
}

// SOAPHeader represents the SOAP header
type SOAPHeader struct {
	Content interface{} `xml:",omitempty"`
}

// SOAPBody represents the SOAP body
type SOAPBody struct {
	Content interface{} `xml:",omitempty"`
	Fault   *SOAPFault  `xml:"Fault,omitempty"`
}

// SOAPFault represents a SOAP fault
type SOAPFault struct {
	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

// SOAPClient implements the Client interface for SOAP APIs
type SOAPClient struct {
	endpoint   Endpoint
	httpClient *http.Client
	config     *APICallerConfig
}

// NewSOAPClient creates a new SOAP client
func NewSOAPClient(endpoint Endpoint, config *APICallerConfig) (*SOAPClient, error) {
	client := &SOAPClient{
		endpoint: endpoint,
		config:   config,
	}

	// Create HTTP client with TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	// Load CA certificate if provided
	if config.CACertPath != "" {
		caCert, err := os.ReadFile(config.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate if provided
	if config.TLSCertPath != "" && config.TLSKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(config.TLSCertPath, config.TLSKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	client.httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 30 * time.Second,
	}

	return client, nil
}

// Call makes a SOAP API call
func (sc *SOAPClient) Call(options *CallOptions) (*Response, error) {
	// SOAP always uses HTTP POST
	httpMethod := "POST"

	// Prepare SOAP envelope
	var bodyBytes []byte
	var err error

	// Generate SOAP envelope automatically from Body parameter
	switch v := options.Body.(type) {
	case []byte:
		// If body is already serialized XML (legacy support), use it directly
		bodyBytes = v
	case string:
		// If body is a string (legacy support), use it directly
		bodyBytes = []byte(v)
	case map[string]interface{}:
		// Generate SOAP envelope from parameter map
		bodyBytes, err = sc.generateSOAPEnvelope(options.Method, v)
		if err != nil {
			return nil, fmt.Errorf("failed to generate SOAP envelope: %w", err)
		}
	case *SOAPEnvelope:
		// If body is already a SOAP envelope (legacy support), marshal it
		bodyBytes, err = xml.MarshalIndent(v, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal SOAP envelope: %w", err)
		}
		// Add XML declaration
		xmlDeclaration := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
		bodyBytes = append(xmlDeclaration, bodyBytes...)
	default:
		// Wrap any other type in a SOAP envelope
		soapEnvelope := &SOAPEnvelope{
			Body: SOAPBody{
				Content: options.Body,
			},
		}
		bodyBytes, err = xml.MarshalIndent(soapEnvelope, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal SOAP request: %w", err)
		}
		// Add XML declaration
		xmlDeclaration := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
		bodyBytes = append(xmlDeclaration, bodyBytes...)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(options.Context, httpMethod, sc.endpoint.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %w", err)
	}

	// Set SOAP-specific headers
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	// Set SOAPAction header if provided in options
	if options.Headers != nil {
		for key, value := range options.Headers {
			req.Header.Set(key, value)
		}
	}

	// Ensure SOAPAction header is set (required by SOAP 1.1)
	if req.Header.Get("SOAPAction") == "" && options.Method != "" {
		req.Header.Set("SOAPAction", fmt.Sprintf(`"%s"`, options.Method))
	}

	// Make the request
	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return &Response{
			Error: fmt.Errorf("SOAP request failed: %w", err),
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Error:      fmt.Errorf("failed to read SOAP response body: %w", err),
		}, err
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}

	// Check for SOAP faults
	var soapResponse SOAPEnvelope
	if err := xml.Unmarshal(respBody, &soapResponse); err == nil {
		if soapResponse.Body.Fault != nil {
			response.Error = fmt.Errorf("SOAP fault: [%s] %s - %s",
				soapResponse.Body.Fault.Code,
				soapResponse.Body.Fault.String,
				soapResponse.Body.Fault.Detail)
			return response, response.Error
		}
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		response.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return response, response.Error
}

// Close closes the SOAP client
func (sc *SOAPClient) Close() error {
	// HTTP client doesn't need explicit closing
	if sc.httpClient != nil {
		sc.httpClient.CloseIdleConnections()
	}
	return nil
}

// generateSOAPEnvelope creates a SOAP envelope from a parameter map
func (sc *SOAPClient) generateSOAPEnvelope(operationName string, params map[string]interface{}) ([]byte, error) {
	// Determine namespace from WSDL URL or use default
	namespace := "http://tempuri.org/"
	if sc.endpoint.WSDLUrl != "" {
		// Extract namespace from WSDL URL (simplified - could be enhanced with actual WSDL parsing)
		// For now, use a heuristic based on the URL
		namespace = extractNamespaceFromURL(sc.endpoint.WSDLUrl)
	}

	// Build SOAP body content XML
	var bodyContent bytes.Buffer
	bodyContent.WriteString(fmt.Sprintf("<%s xmlns=\"%s\">", operationName, namespace))

	// Add parameters as XML elements
	for key, value := range params {
		bodyContent.WriteString(fmt.Sprintf("<%s>%v</%s>", key, value, key))
	}

	bodyContent.WriteString(fmt.Sprintf("</%s>", operationName))

	// Create complete SOAP envelope
	envelope := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    %s
  </soap:Body>
</soap:Envelope>`, bodyContent.String())

	return []byte(envelope), nil
}

// extractNamespaceFromURL attempts to extract the namespace from the WSDL URL
func extractNamespaceFromURL(wsdlURL string) string {
	// Remove ?wsdl or ?WSDL suffix
	url := wsdlURL
	if len(url) > 5 && (url[len(url)-5:] == "?wsdl" || url[len(url)-5:] == "?WSDL") {
		url = url[:len(url)-5]
	}

	// Remove the filename (e.g., tempconvert.asmx) to get the directory
	lastSlash := -1
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash > 0 {
		url = url[:lastSlash+1]
	} else {
		// Fallback: just add trailing slash
		url = url + "/"
	}

	return url
}

// CreateSOAPEnvelope is a helper function to create a SOAP envelope with the given body content
// Deprecated: Use map[string]interface{} parameters instead for automatic envelope generation
func CreateSOAPEnvelope(bodyContent interface{}, header interface{}) *SOAPEnvelope {
	envelope := &SOAPEnvelope{
		Body: SOAPBody{
			Content: bodyContent,
		},
	}

	if header != nil {
		envelope.Header = &SOAPHeader{
			Content: header,
		}
	}

	return envelope
}
