package cinema

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/kazalab/zaphub/internal/model"
)

const allocineBaseURL = "https://www.allocine.fr"

type AllocineProvider struct {
	url    string
	client *http.Client
}

func NewAllocineProvider(url string, client *http.Client) *AllocineProvider {
	if client == nil {
		client = defaultClient()
	}
	return &AllocineProvider{url: url, client: client}
}

func (p *AllocineProvider) Name() string { return "allocine" }

func (p *AllocineProvider) Fetch(ctx context.Context) ([]model.CinemaRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil, fmt.Errorf("cinema request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126 Safari/537.36")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	req.Header.Set("Accept", "text/html")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cinema download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cinema download: status %s", resp.Status)
	}

	return Parse(resp.Body)
}

func Parse(r io.Reader) ([]model.CinemaRelease, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("html parse: %w", err)
	}

	var releases []model.CinemaRelease
	var current *model.CinemaRelease

	walk(doc, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		switch {
		case isCard(n):
			if current != nil {
				*current = finishRelease(*current, n)
				releases = append(releases, *current)
			}
			rel := parseCard(n)
			current = &rel
		case hasClass(n, "rating-holder") && current != nil && current.PressRating == 0 && current.SpectatorRating == 0:
			current.PressRating, current.SpectatorRating = parseRatings(n)
		case hasClass(n, "synopsis") && current != nil && current.Synopsis == "":
			current.Synopsis = parseSynopsis(n)
		}
	})

	if current != nil {
		*current = finishRelease(*current, nil)
		releases = append(releases, *current)
	}

	return releases, nil
}

func parseCard(card *html.Node) model.CinemaRelease {
	rel := model.CinemaRelease{}

	poster := findDescByClass(card, "thumbnail-img")
	if poster != nil {
		src := getAttr(poster, "data-src")
		if src == "" {
			src = getAttr(poster, "src")
		}
		if src != "" && !strings.HasPrefix(src, "data:") {
			rel.Poster = src
		}
	}

	if a := findDescByClass(card, "meta-title-link"); a != nil {
		rel.Title = nodeText(a)
		rel.Link = makeAbsURL(getAttr(a, "href"))
	}

	badge := findDescByClass(card, "js-affinity-badge")
	if badge != nil {
		rel.ID = getAttr(badge, "data-entity-id")
	}

	info := findDescByClass(card, "meta-body-info")
	if info != nil {
		rel.ReleaseDate, rel.Duration, rel.DurationMinutes, rel.Genres = parseInfo(info)
	}

	dir := findDescByClass(card, "meta-body-direction")
	if dir != nil {
		rel.Director = parseDirection(dir)
	}

	actor := findDescByClass(card, "meta-body-actor")
	if actor != nil {
		rel.Actors = parseActors(actor)
	}

	for _, item := range findDescsByClass(card, "meta-body-item") {
		if hasClassExact(item, "meta-body-info") || hasClassExact(item, "meta-body-direction") || hasClassExact(item, "meta-body-actor") || hasClassExact(item, "gelule-holder") {
			continue
		}
		if light := findDescByClass(item, "light"); light != nil && strings.Contains(nodeText(light), "Titre original") {
			if val := findDescByClass(item, "dark-grey"); val != nil {
				rel.OriginalTitle = strings.TrimSpace(nodeText(val))
			}
		}
	}

	return rel
}

func parseInfo(info *html.Node) (date, duration string, durationMinutes int, genres []string) {
	for _, span := range findDescsByClass(info, "date") {
		date = nodeText(span)
	}

	flat := flattenText(info)
	parts := strings.Split(flat, "|")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if isDuration(p) {
			duration = p
			durationMinutes = parseDuration(p)
		} else if p != date && !isDuration(p) && !isAgeRating(p) {
			for _, g := range strings.Split(p, ",") {
				g = strings.TrimSpace(g)
				if g != "" {
					genres = append(genres, g)
				}
			}
		}
	}

	return
}

func parseDirection(dir *html.Node) string {
	for _, span := range findDescsByClass(dir, "light") {
		if nodeText(span) == "De" {
			return extractAfterLabel(span)
		}
	}
	return ""
}

func parseActors(act *html.Node) []string {
	for _, span := range findDescsByClass(act, "light") {
		if nodeText(span) == "Avec" {
			sibling := span.NextSibling
			var actors []string
			for sibling != nil {
				if sibling.Type == html.ElementNode {
					t := strings.TrimSpace(nodeText(sibling))
					if t != "" {
						actors = append(actors, t)
					}
				}
				sibling = sibling.NextSibling
			}
			return actors
		}
	}
	return nil
}

func extractAfterLabel(label *html.Node) string {
	sibling := label.NextSibling
	for sibling != nil {
		if sibling.Type == html.ElementNode {
			t := strings.TrimSpace(nodeText(sibling))
			if t != "" {
				return t
			}
		}
		if sibling.Type == html.TextNode {
			t := strings.TrimSpace(sibling.Data)
			if t != "" && t != "," {
				return t
			}
		}
		sibling = sibling.NextSibling
	}
	return ""
}

func parseRatings(rh *html.Node) (pressRating, spectatorRating float64) {
	for _, item := range findDescsByClass(rh, "rating-item-content") {
		if hasClass(item, "js-user-friends-rating") {
			continue
		}
		title := findDescByClass(item, "rating-title")
		note := findDescByClass(item, "stareval-note")
		if title == nil || note == nil || hasClass(note, "no-rating") {
			continue
		}
		val := strings.TrimSpace(nodeText(note))
		val = strings.Replace(val, ",", ".", 1)
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			continue
		}
		if strings.Contains(nodeText(title), "Press") {
			pressRating = f
		} else if strings.Contains(nodeText(title), "Spectat") {
			spectatorRating = f
		}
	}
	return
}

func parseSynopsis(syn *html.Node) string {
	ct := findDescByClass(syn, "content-txt")
	if ct == nil {
		return ""
	}
	return strings.TrimSpace(flattenText(ct))
}

func finishRelease(rel model.CinemaRelease, _ *html.Node) model.CinemaRelease {
	return rel
}

// --- HTML helpers ---

func walk(n *html.Node, fn func(*html.Node)) {
	fn(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, fn)
	}
}

func hasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(getAttr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

func hasClassExact(n *html.Node, class string) bool {
	for _, c := range strings.Fields(getAttr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

func isCard(n *html.Node) bool {
	return n.DataAtom == atom.Div && hasClass(n, "entity-card") && hasClass(n, "entity-card-list")
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func findDescByClass(n *html.Node, class string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && hasClass(c, class) {
			return c
		}
		if found := findDescByClass(c, class); found != nil {
			return found
		}
	}
	return nil
}

func findDescsByClass(n *html.Node, class string) []*html.Node {
	var result []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && hasClass(c, class) {
			result = append(result, c)
		}
		result = append(result, findDescsByClass(c, class)...)
	}
	return result
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walkText func(*html.Node)
	walkText = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkText(c)
		}
	}
	walkText(n)
	return b.String()
}

func flattenText(n *html.Node) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, nodeText(n))
}

func isDuration(s string) bool {
	return strings.Contains(s, "h") && strings.Contains(s, "min")
}

func isAgeRating(s string) bool {
	return strings.HasPrefix(s, "Dès") || strings.HasPrefix(s, "Interdit")
}

func parseDuration(s string) int {
	s = strings.TrimSpace(s)
	var hours, mins int
	_, _ = fmt.Sscanf(s, "%dh %dmin", &hours, &mins)
	return hours*60 + mins
}

func makeAbsURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http") {
		return path
	}
	return allocineBaseURL + path
}
