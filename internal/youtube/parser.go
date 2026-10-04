package youtube

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const xmlTimeLayout = time.RFC3339

type atomFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Title   string       `xml:"title"`
	Author  []atomAuthor `xml:"author"`
	Entries []atomEntry  `xml:"entry"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomEntry struct {
	ID        string     `xml:"id"`
	Title     string     `xml:"title"`
	Link      []atomLink `xml:"link"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	Media     atomMedia  `xml:"group"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type atomMedia struct {
	Title       string        `xml:"title"`
	Description string        `xml:"description"`
	Thumbnails  []atomThumb   `xml:"thumbnail"`
	Duration    string        `xml:"duration"`
	Community   atomCommunity `xml:"community"`
}

type atomThumb struct {
	URL string `xml:"url,attr"`
}

type atomCommunity struct {
	StarRating atomStarRating `xml:"starRating"`
}

type atomStarRating struct {
	Count string `xml:"count,attr"`
}

type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Published   time.Time `json:"published"`
	Updated     time.Time `json:"updated"`
	Thumbnail   string    `json:"thumbnail"`
	Likes       int       `json:"likes"`
	Duration    string    `json:"duration"`
	ChannelID   string    `json:"channel_id"`
	ChannelName string    `json:"channel_name"`
}

func Parse(r io.Reader, channelID, channelName string) ([]Video, error) {
	dec := xml.NewDecoder(r)
	var feed atomFeed
	if err := dec.Decode(&feed); err != nil {
		return nil, fmt.Errorf("atom decode: %w", err)
	}

	name := strings.TrimSpace(channelName)
	if name == "" && len(feed.Author) > 0 {
		name = strings.TrimSpace(feed.Author[0].Name)
	}
	if name == "" {
		name = strings.TrimSpace(feed.Title)
	}

	videos := make([]Video, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		vid := strings.TrimPrefix(strings.TrimSpace(e.ID), "yt:video:")
		link := ""
		for _, l := range e.Link {
			if l.Href != "" {
				link = l.Href
				break
			}
		}
		published, _ := time.Parse(xmlTimeLayout, strings.TrimSpace(e.Published))
		updated, _ := time.Parse(xmlTimeLayout, strings.TrimSpace(e.Updated))
		thumbnail := ""
		if len(e.Media.Thumbnails) > 0 {
			thumbnail = e.Media.Thumbnails[0].URL
		}
		likes, _ := strconv.Atoi(strings.TrimSpace(e.Media.Community.StarRating.Count))
		videos = append(videos, Video{
			ID:          vid,
			Title:       strings.TrimSpace(e.Title),
			Description: strings.TrimSpace(e.Media.Description),
			Link:        link,
			Published:   published,
			Updated:     updated,
			Thumbnail:   thumbnail,
			Likes:       likes,
			Duration:    strings.TrimSpace(e.Media.Duration),
			ChannelID:   channelID,
			ChannelName: name,
		})
	}
	return videos, nil
}

func ParseBytes(data []byte, channelID, channelName string) ([]Video, error) {
	return Parse(bytes.NewReader(data), channelID, channelName)
}
