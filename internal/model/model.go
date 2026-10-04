package model

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
}

type Programme struct {
	Title       string   `json:"title"`
	SubTitle    string   `json:"sub_title,omitempty"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Start       string   `json:"start"`
	Stop        string   `json:"stop"`
	Categories  []string `json:"categories,omitempty"`
	Rating      string   `json:"rating,omitempty"`
}

type Video struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Link        string `json:"link"`
	Published   string `json:"published"`
	Updated     string `json:"updated"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	Likes       int    `json:"likes,omitempty"`
	Duration    string `json:"duration,omitempty"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
}

type EveningProgramme struct {
	Title       string   `json:"title"`
	SubTitle    string   `json:"sub_title,omitempty"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Start       string   `json:"start"`
	Stop        string   `json:"stop"`
	Categories  []string `json:"categories,omitempty"`
	Rating      string   `json:"rating,omitempty"`
}

type EveningChannel struct {
	ChannelID   string             `json:"channel_id"`
	ChannelName string             `json:"channel_name"`
	Icon        string             `json:"icon,omitempty"`
	Programmes  []EveningProgramme `json:"programmes"`
}

type EveningResponse struct {
	Date     string           `json:"date"`
	Channels []EveningChannel `json:"channels"`
}

type CinemaRelease struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	OriginalTitle   string   `json:"original_title,omitempty"`
	Poster          string   `json:"poster,omitempty"`
	ReleaseDate     string   `json:"release_date,omitempty"`
	Duration        string   `json:"duration,omitempty"`
	DurationMinutes int      `json:"duration_minutes,omitempty"`
	Genres          []string `json:"genres,omitempty"`
	Director        string   `json:"director,omitempty"`
	Actors          []string `json:"actors,omitempty"`
	Synopsis        string   `json:"synopsis,omitempty"`
	PressRating     float64  `json:"press_rating,omitempty"`
	SpectatorRating float64  `json:"spectator_rating,omitempty"`
	Link            string   `json:"link,omitempty"`
}
