package epg

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kazalab/zaphub/internal/model"
)

type Fetcher struct {
	url          string
	client       *http.Client
	lastModified string
}

func NewFetcher(url string, client *http.Client) *Fetcher {
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	return &Fetcher{url: url, client: client}
}

// Fetch retrieves and parses the XMLTV file. It returns changed=false when the
// upstream answered 304 Not Modified (data unchanged since last fetch).
func (f *Fetcher) Fetch(ctx context.Context) ([]model.Channel, []Programme, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, nil, false, fmt.Errorf("epg request: %w", err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	if f.lastModified != "" {
		req.Header.Set("If-Modified-Since", f.lastModified)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, nil, false, fmt.Errorf("epg download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return nil, nil, false, nil
	case http.StatusOK:
	default:
		return nil, nil, false, fmt.Errorf("epg download: status %s", resp.Status)
	}

	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		f.lastModified = lm
	}

	body, err := decodeBody(resp)
	if err != nil {
		return nil, nil, false, err
	}
	defer func() { _ = body.Close() }()

	channels, programmes, err := Parse(body)
	if err != nil {
		return nil, nil, false, err
	}
	return channels, programmes, true, nil
}

func decodeBody(resp *http.Response) (io.ReadCloser, error) {
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip open: %w", err)
		}
		return gz, nil
	}
	br := bufio.NewReader(resp.Body)
	if magic, err := br.Peek(2); err == nil && len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		gz, err := gzip.NewReader(br)
		if err != nil {
			return nil, fmt.Errorf("gzip open: %w", err)
		}
		return gz, nil
	}
	return io.NopCloser(br), nil
}
