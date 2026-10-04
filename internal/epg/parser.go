package epg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kazalab/zaphub/internal/model"
)

const xmlTimeLayout = "20060102150405 -0700"

type xmltv struct {
	XMLName    xml.Name       `xml:"tv"`
	Channels   []xmlChannel   `xml:"channel"`
	Programmes []xmlProgramme `xml:"programme"`
}

type xmlChannel struct {
	ID          string    `xml:"id,attr"`
	DisplayName []string  `xml:"display-name"`
	Icons       []xmlIcon `xml:"icon"`
}

type xmlIcon struct {
	Src string `xml:"src,attr"`
}

type xmlProgramme struct {
	Start      string      `xml:"start,attr"`
	Stop       string      `xml:"stop,attr"`
	Channel    string      `xml:"channel,attr"`
	Titles     []xmlText   `xml:"title"`
	SubTitles  []xmlText   `xml:"sub-title"`
	Descs      []xmlText   `xml:"desc"`
	Icons      []xmlIcon   `xml:"icon"`
	Categories []xmlText   `xml:"category"`
	Ratings    []xmlRating `xml:"rating"`
}

type xmlText struct {
	Text string `xml:",chardata"`
}

type xmlRating struct {
	Values []xmlText `xml:"value"`
}

type Programme struct {
	Channel    string
	Start      time.Time
	Stop       time.Time
	Title      string
	SubTitle   string
	Desc       string
	Icon       string
	Categories []string
	Rating     string
}

func Parse(r io.Reader) ([]model.Channel, []Programme, error) {
	dec := xml.NewDecoder(r)
	var tv xmltv
	if err := dec.Decode(&tv); err != nil {
		return nil, nil, fmt.Errorf("xmltv decode: %w", err)
	}

	channels := make([]model.Channel, 0, len(tv.Channels))
	for _, c := range tv.Channels {
		ch := model.Channel{ID: c.ID, Name: firstText(c.DisplayName)}
		if len(c.Icons) > 0 {
			ch.Icon = c.Icons[0].Src
		}
		channels = append(channels, ch)
	}

	programmes := make([]Programme, 0, len(tv.Programmes))
	for _, p := range tv.Programmes {
		start, err := time.Parse(xmlTimeLayout, strings.TrimSpace(p.Start))
		if err != nil {
			return nil, nil, fmt.Errorf("programme %q start %q: %w", p.Channel, p.Start, err)
		}
		stop, err := time.Parse(xmlTimeLayout, strings.TrimSpace(p.Stop))
		if err != nil {
			return nil, nil, fmt.Errorf("programme %q stop %q: %w", p.Channel, p.Stop, err)
		}
		prog := Programme{
			Channel:  p.Channel,
			Start:    start,
			Stop:     stop,
			Title:    textOf(p.Titles),
			SubTitle: textOf(p.SubTitles),
			Desc:     textOf(p.Descs),
		}
		if len(p.Icons) > 0 {
			prog.Icon = p.Icons[0].Src
		}
		for _, cat := range p.Categories {
			if s := strings.TrimSpace(cat.Text); s != "" {
				prog.Categories = append(prog.Categories, s)
			}
		}
		if len(p.Ratings) > 0 && len(p.Ratings[0].Values) > 0 {
			prog.Rating = strings.TrimSpace(p.Ratings[0].Values[0].Text)
		}
		programmes = append(programmes, prog)
	}
	return channels, programmes, nil
}

func firstText(xs []string) string {
	for _, x := range xs {
		if s := strings.TrimSpace(x); s != "" {
			return s
		}
	}
	return ""
}

func textOf(xs []xmlText) string {
	for _, x := range xs {
		if s := strings.TrimSpace(x.Text); s != "" {
			return s
		}
	}
	return ""
}
