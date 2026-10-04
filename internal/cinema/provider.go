package cinema

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kazalab/zaphub/internal/model"
)

type Provider interface {
	Name() string
	Fetch(ctx context.Context) ([]model.CinemaRelease, error)
}

func NewProvider(kind, url string, client *http.Client) (Provider, error) {
	switch kind {
	case "allocine":
		return NewAllocineProvider(url, client), nil
	default:
		return nil, fmt.Errorf("unknown cinema provider: %q", kind)
	}
}

func defaultClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}
