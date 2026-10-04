package youtube

import (
	"context"
	"net/http"
)

type Refresher struct {
	client *http.Client
	store  *Store
}

func NewRefresher(client *http.Client, store *Store) *Refresher {
	return &Refresher{client: client, store: store}
}

func (r *Refresher) Refresh(ctx context.Context) {
	for _, ch := range r.store.Channels() {
		r.RefreshChannel(ctx, ch)
	}
}

func (r *Refresher) RefreshChannel(ctx context.Context, ch Channel) {
	videos, err := FetchChannel(ctx, r.client, ch)
	if err != nil {
		r.store.SetError(err)
		return
	}
	if len(videos) > 0 {
		name := videos[0].ChannelName
		if name != "" && name != ch.ID {
			ch.Name = name
			_ = r.store.SetChannel(ch)
		}
	}
	r.store.UpdateVideos(ch.ID, videos)
}
