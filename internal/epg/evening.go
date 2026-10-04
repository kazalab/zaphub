package epg

import (
	"sort"
	"time"
	_ "time/tzdata"
)

const (
	// EveningStartMin / EveningStartMax define the inclusive start-time window
	// of the evening slot, expressed in minutes since midnight (20:50 / 23:50).
	EveningStartMin = 20*60 + 50
	EveningStartMax = 23*60 + 50
	// EveningMinDuration is the minimum programme length to keep.
	EveningMinDuration = 40 * time.Minute
)

var paris = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return time.FixedZone("CET", 3600)
	}
	return loc
}()

// EveningProgrammes returns, per channel, all programmes starting on the given
// date within the evening slot (start between 20:50 and 23:50 local time) and
// lasting at least EveningMinDuration, sorted by start time.
func EveningProgrammes(programmes []Programme, date time.Time) map[string][]Programme {
	target := date.In(paris)
	day := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, paris)

	grouped := make(map[string][]Programme)
	for _, p := range programmes {
		start := p.Start.In(paris)
		if start.Before(day) || start.YearDay() != day.YearDay() {
			continue
		}
		minutes := start.Hour()*60 + start.Minute()
		if minutes < EveningStartMin || minutes > EveningStartMax {
			continue
		}
		if p.Stop.Sub(p.Start) < EveningMinDuration {
			continue
		}
		grouped[p.Channel] = append(grouped[p.Channel], p)
	}
	for ch := range grouped {
		sort.Slice(grouped[ch], func(i, j int) bool { return grouped[ch][i].Start.Before(grouped[ch][j].Start) })
	}
	return grouped
}
