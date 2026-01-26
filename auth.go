package tradera

import (
	"bytes"
	"encoding/xml"
)

const (
	// TraderaNamespace is the XML namespace for Tradera SOAP headers.
	TraderaNamespace = "http://api.tradera.com"

	// SOAP namespaces
	SOAPEnvelopeNamespace = "http://schemas.xmlsoap.org/soap/envelope/"
)

// AuthenticationHeader is the SOAP header for app authentication.
// This header is required for all API calls.
type AuthenticationHeader struct {
	XMLName xml.Name `xml:"tra:AuthenticationHeader"`
	AppID   int      `xml:"tra:AppId"`
	AppKey  string   `xml:"tra:AppKey"`
}

// AuthorizationHeader is the SOAP header for user authentication.
// This header is required for Restricted, Order, and Buyer services.
type AuthorizationHeader struct {
	XMLName xml.Name `xml:"tra:AuthorizationHeader"`
	UserID  int      `xml:"tra:UserId"`
	Token   string   `xml:"tra:Token"`
}

// SOAPHeaders contains all headers to be included in a SOAP request.
type SOAPHeaders struct {
	Authentication AuthenticationHeader
	Authorization  *AuthorizationHeader // nil if not using user auth
}

// NewSOAPHeaders creates SOAP headers from the config.
func NewSOAPHeaders(cfg Config) SOAPHeaders {
	headers := SOAPHeaders{
		Authentication: AuthenticationHeader{
			AppID:  cfg.AppID,
			AppKey: cfg.AppKey,
		},
	}

	if cfg.HasUserAuth() {
		headers.Authorization = &AuthorizationHeader{
			UserID: cfg.UserID,
			Token:  cfg.Token,
		}
	}

	return headers
}

// MarshalXML implements xml.Marshaler for SOAPHeaders.
func (h SOAPHeaders) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Marshal AuthenticationHeader
	if err := e.Encode(h.Authentication); err != nil {
		return err
	}

	// Marshal AuthorizationHeader if present
	if h.Authorization != nil {
		if err := e.Encode(h.Authorization); err != nil {
			return err
		}
	}

	return nil
}

// BuildSOAPHeaderXML creates the XML bytes for SOAP headers.
func BuildSOAPHeaderXML(cfg Config) ([]byte, error) {
	headers := NewSOAPHeaders(cfg)

	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	if err := encoder.Encode(headers.Authentication); err != nil {
		return nil, err
	}

	if headers.Authorization != nil {
		if err := encoder.Encode(headers.Authorization); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// RequireUserAuth is a helper that returns an error if user auth is not configured.
func RequireUserAuth(cfg Config) error {
	if !cfg.HasUserAuth() {
		return ErrAuthRequired
	}
	return nil
}
