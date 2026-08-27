package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
)

// SanitizingTransport wraps an http.RoundTripper and strips XML-illegal control characters
// from response bodies before they reach a SOAP/XML decoder. Tradera's SOAP responses
// occasionally contain raw control characters (observed in practice: U+0002, U+0008, U+001A)
// inside item title/description text - almost certainly garbage bytes in seller-entered
// content - which Go's encoding/xml rejects outright, failing the entire response rather than
// just the offending field.
//
// Per the XML 1.0 spec (https://www.w3.org/TR/xml/#charsets), the only legal characters below
// U+0020 are tab (U+0009), line feed (U+000A), and carriage return (U+000D); every other
// control character has no valid interpretation in XML, so removing them cannot discard any
// data XML could have represented in the first place.
type SanitizingTransport struct {
	Base http.RoundTripper
}

// NewSanitizingTransport wraps base (or http.DefaultTransport if base is nil) with response
// body sanitization.
func NewSanitizingTransport(base http.RoundTripper) *SanitizingTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &SanitizingTransport{Base: base}
}

// RoundTrip implements http.RoundTripper.
func (t *SanitizingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.Base.RoundTrip(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}

	sanitized := StripIllegalXMLChars(body)
	resp.Body = io.NopCloser(bytes.NewReader(sanitized))
	resp.ContentLength = int64(len(sanitized))
	if resp.Header != nil {
		resp.Header.Set("Content-Length", strconv.Itoa(len(sanitized)))
	}

	return resp, nil
}

// StripIllegalXMLChars removes bytes that are illegal anywhere in an XML 1.0 document: control
// characters below U+0020 other than tab, line feed, and carriage return. Only single-byte
// ASCII control characters are targeted - a byte in that range can never be part of a
// multi-byte UTF-8 sequence (those always use continuation bytes in 0x80-0xBF), so this cannot
// corrupt valid UTF-8 content elsewhere in the body.
func StripIllegalXMLChars(body []byte) []byte {
	hasIllegal := false
	for _, b := range body {
		if isIllegalXMLByte(b) {
			hasIllegal = true
			break
		}
	}
	if !hasIllegal {
		return body
	}

	cleaned := make([]byte, 0, len(body))
	for _, b := range body {
		if isIllegalXMLByte(b) {
			continue
		}
		cleaned = append(cleaned, b)
	}
	return cleaned
}

func isIllegalXMLByte(b byte) bool {
	return b < 0x20 && b != 0x09 && b != 0x0A && b != 0x0D
}
