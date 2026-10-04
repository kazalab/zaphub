package epg

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFetcher(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "epg_sample.xml"))
	if err != nil {
		t.Fatal(err)
	}
	gz := gzipBytes(t, raw)

	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("If-Modified-Since") != "" {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Last-Modified", "Wed, 12 Aug 2026 02:10:00 GMT")
		if _, err := w.Write(gz); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer srv.Close()

	f := NewFetcher(srv.URL, srv.Client())

	channels, programmes, changed, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true on first fetch")
	}
	if len(channels) != 3 || len(programmes) != 6 {
		t.Fatalf("unexpected parse result: %d channels, %d programmes", len(channels), len(programmes))
	}

	_, _, changed, err = f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false on 304")
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}
