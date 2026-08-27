package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStripIllegalXMLChars(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want []byte
	}{
		{"no illegal chars", []byte("<a>hello</a>"), []byte("<a>hello</a>")},
		{"strips SUB (U+001A)", []byte("<a>hel\x1Alo</a>"), []byte("<a>hello</a>")},
		{"strips multiple control chars", []byte("<a>a\x02b\x08c</a>"), []byte("<a>abc</a>")},
		{"keeps tab, LF, CR", []byte("<a>a\tb\nc\rd</a>"), []byte("<a>a\tb\nc\rd</a>")},
		{"keeps multi-byte UTF-8", []byte("<a>Bröllopsslöjor</a>"), []byte("<a>Bröllopsslöjor</a>")},
		{"empty body", []byte{}, []byte{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StripIllegalXMLChars(tc.in)
			if !bytes.Equal(got, tc.want) {
				t.Errorf("StripIllegalXMLChars(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSanitizingTransport_RoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<a>hel\x1Alo</a>"))
	}))
	defer server.Close()

	client := &http.Client{Transport: NewSanitizingTransport(nil)}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	want := "<a>hello</a>"
	if string(body) != want {
		t.Errorf("body = %q, want %q", body, want)
	}
	if resp.ContentLength != int64(len(want)) {
		t.Errorf("ContentLength = %d, want %d", resp.ContentLength, len(want))
	}
	if got := resp.Header.Get("Content-Length"); got != "12" {
		t.Errorf("Content-Length header = %q, want %q", got, "12")
	}
}

func TestSanitizingTransport_NoIllegalChars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<a>clean</a>"))
	}))
	defer server.Close()

	client := &http.Client{Transport: NewSanitizingTransport(nil)}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body) != "<a>clean</a>" {
		t.Errorf("body = %q, want unchanged", body)
	}
}

func TestSanitizingTransport_PropagatesTransportError(t *testing.T) {
	transport := NewSanitizingTransport(roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	}))
	client := &http.Client{Transport: transport}

	_, err := client.Get("http://example.invalid")
	if err == nil || !strings.Contains(err.Error(), io.ErrUnexpectedEOF.Error()) {
		t.Errorf("expected underlying transport error to propagate, got: %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
