package middleware

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

// SanitizingTransport wraps an http.RoundTripper and strips content that has no valid
// interpretation in an XML 1.0 document from response bodies, before they reach a SOAP/XML
// decoder. Tradera's SOAP responses have been observed to contain numeric character
// references (e.g. "&#2;", "&#x1A;") inside item title/description text that decode to
// illegal XML characters - almost certainly garbage bytes in seller-entered content that got
// entity-encoded somewhere in Tradera's own pipeline. The wire bytes themselves are
// well-formed ASCII, so this can only be caught by decoding character references and
// checking the resulting codepoint, not by scanning raw bytes. Go's encoding/xml rejects the
// whole response on a single bad reference rather than just the offending field.
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

// numericCharRefPattern matches XML numeric character references: &#nnnn; and &#xhhhh;
var numericCharRefPattern = regexp.MustCompile(`&#(x[0-9a-fA-F]+|[0-9]+);`)

// StripIllegalXMLChars removes content that has no valid interpretation in an XML 1.0
// document:
//
//  1. Numeric character references (&#2; or &#x1A;) that decode to an illegal codepoint -
//     the reference itself is dropped, since Go's encoding/xml rejects the whole response on
//     decoding one, even though the raw wire bytes contain nothing illegal.
//  2. Raw control-character bytes below U+0020 other than tab, line feed, and carriage
//     return, kept as a defense-in-depth measure in case an illegal character ever appears
//     unencoded rather than as a reference.
func StripIllegalXMLChars(body []byte) []byte {
	body = numericCharRefPattern.ReplaceAllFunc(body, func(match []byte) []byte {
		inner := match[2 : len(match)-1] // strip leading "&#" and trailing ";"

		var value int64
		var err error
		if len(inner) > 0 && (inner[0] == 'x' || inner[0] == 'X') {
			value, err = strconv.ParseInt(string(inner[1:]), 16, 32)
		} else {
			value, err = strconv.ParseInt(string(inner), 10, 32)
		}
		if err != nil || isIllegalXMLCodepoint(rune(value)) {
			return nil
		}
		return match
	})

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

// isIllegalXMLCodepoint reports whether r has no valid interpretation in an XML 1.0 document,
// per https://www.w3.org/TR/xml/#charsets:
//
//	Char ::= #x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | [#x10000-#x10FFFF]
func isIllegalXMLCodepoint(r rune) bool {
	switch {
	case r == 0x09 || r == 0x0A || r == 0x0D:
		return false
	case r < 0x20:
		return true
	case r >= 0xD800 && r <= 0xDFFF: // surrogate range, invalid outside UTF-16 encoding
		return true
	case r == 0xFFFE || r == 0xFFFF:
		return true
	case r > 0x10FFFF:
		return true
	default:
		return false
	}
}
