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
		{"no illegal content", []byte("<a>hello</a>"), []byte("<a>hello</a>")},
		{"strips raw SUB byte (U+001A)", []byte("<a>hel\x1Alo</a>"), []byte("<a>hello</a>")},
		{"strips multiple raw control bytes", []byte("<a>a\x02b\x08c</a>"), []byte("<a>abc</a>")},
		{"keeps raw tab, LF, CR", []byte("<a>a\tb\nc\rd</a>"), []byte("<a>a\tb\nc\rd</a>")},
		{"keeps multi-byte UTF-8", []byte("<a>Bröllopsslöjor</a>"), []byte("<a>Bröllopsslöjor</a>")},
		{"empty body", []byte{}, []byte{}},

		// The actual bug: a well-formed-ASCII numeric character reference that decodes to an
		// illegal codepoint. Raw byte scanning alone never catches this - the wire bytes are
		// just '&', '#', digits, ';'.
		{"strips decimal numeric ref to illegal codepoint (&#2;)", []byte("<a>hel&#2;lo</a>"), []byte("<a>hello</a>")},
		{"strips hex numeric ref to illegal codepoint (&#x1A;)", []byte("<a>hel&#x1A;lo</a>"), []byte("<a>hello</a>")},
		{"strips hex numeric ref, lowercase x and digits (&#x8;)", []byte("<a>a&#x8;b</a>"), []byte("<a>ab</a>")},
		{"keeps a legal numeric ref (&#65; = 'A')", []byte("<a>&#65;</a>"), []byte("<a>&#65;</a>")},
		{"keeps a legal hex numeric ref (&#x41; = 'A')", []byte("<a>&#x41;</a>"), []byte("<a>&#x41;</a>")},
		{"keeps numeric ref to tab/LF/CR", []byte("<a>&#9;&#10;&#13;</a>"), []byte("<a>&#9;&#10;&#13;</a>")},
		{"strips numeric ref to surrogate codepoint", []byte("<a>a&#xD800;b</a>"), []byte("<a>ab</a>")},
		{"ignores malformed numeric ref instead of erroring", []byte("<a>&#;</a>"), []byte("<a>&#;</a>")},
		{"multiple illegal refs in one body", []byte("<a>&#2;x&#8;y&#26;</a>"), []byte("<a>xy</a>")},
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
		w.Write([]byte("<a>hel&#x1A;lo</a>"))
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

func TestSanitizingTransport_NoIllegalContent(t *testing.T) {
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
