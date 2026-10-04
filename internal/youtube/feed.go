package youtube

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const feedURL = "https://www.youtube.com/feeds/videos.xml"

func FeedURL(channelID string) string {
	u, _ := url.Parse(feedURL)
	q := u.Query()
	q.Set("channel_id", channelID)
	u.RawQuery = q.Encode()
	return u.String()
}

func FetchChannel(ctx context.Context, client *http.Client, ch Channel) ([]Video, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, FeedURL(ch.ID), nil)
	if err != nil {
		return nil, fmt.Errorf("youtube request: %w", err)
	}
	req.Header.Set("User-Agent", "zaphub/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube download %s: %w", ch.ID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("youtube channel %s: not found", ch.ID)
	default:
		return nil, fmt.Errorf("youtube channel %s: status %s", ch.ID, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("youtube read %s: %w", ch.ID, err)
	}
	videos, err := ParseBytes(body, ch.ID, ch.Name)
	if err != nil {
		return nil, fmt.Errorf("youtube parse %s: %w", ch.ID, err)
	}
	return videos, nil
}
