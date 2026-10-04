package cinema

import (
	"strings"
	"testing"
)

const testHTML = `<html><body>
<div class="card entity-card entity-card-list cf">
  <figure class="thumbnail">
    <span class="thumbnail-container thumbnail-link">
      <img class="thumbnail-img" src="https://fr.web.img2.acsta.net/c_310_420/img/abc.jpg" alt="poster" />
    </span>
  </figure>
  <div class="meta meta-affinity-score">
    <h2 class="meta-title">
      <a class="meta-title-link" href="/film/fichefilm_gen_cfilm=300784.html">La Dernière patiente</a>
    </h2>
    <div class="meta-body">
      <div class="meta-body-item meta-body-info">
        <span class="date">2 septembre 2026</span>
        <span class="spacer">|</span>
        1h 20min
        <span class="spacer">|</span>
        <span class="dark-grey-link">Drame</span>
      </div>
      <div class="meta-body-item meta-body-direction">
        <span class="light">De</span>
        <span class="dark-grey-link">Rémi Bassaler</span>
      </div>
      <div class="meta-body-item meta-body-actor">
        <span class="light">Avec</span>
        <a class="dark-grey-link" href="/personne/fichepersonne_gen_cpersonne=5052.html">Anouk Grinberg</a>,
        <span class="dark-grey-link">Félix Lefebvre</span>
      </div>
      <div class="meta-body-item gelule-holder"></div>
    </div>
    <div class="js-affinity-badge" data-entity-id="TW92aWU6MzAwNzg0"></div>
  </div>
</div>
<div class="rating-holder rating-holder-3">
  <div class="rating-item">
    <div class="rating-item-content">
      <span class="rating-title"> Presse </span>
      <div class="stareval stareval-small stareval-theme-default">
        <span class="stareval-note">2,8</span>
      </div>
    </div>
  </div>
  <div class="rating-item">
    <div class="rating-item-content js-user-friends-rating">
      <span class="rating-title"> Amis </span>
      <div class="stareval stareval-small stareval-theme-default">
        <span class="stareval-note">3,5</span>
      </div>
    </div>
  </div>
  <div class="rating-item">
    <div class="rating-item-content">
      <span class="rating-title"> Spectateurs </span>
      <div class="stareval stareval-small stareval-theme-default">
        <span class="stareval-note">3,4</span>
      </div>
    </div>
  </div>
</div>
<div class="synopsis"><div class="content-txt">Film sur une patiente.</div></div>

<div class="card entity-card entity-card-list cf">
  <figure class="thumbnail">
    <span class="thumbnail-container thumbnail-link">
      <img class="thumbnail-img" data-src="https://fr.web.img5.acsta.net/c_310_420/img/lazy.jpg" alt="poster" />
    </span>
  </figure>
  <div class="meta meta-affinity-score">
    <h2 class="meta-title">
      <a class="meta-title-link" href="/film/fichefilm_gen_cfilm=1000045609.html">Katy Perry: The Lifetimes Tour</a>
    </h2>
    <div class="meta-body">
      <div class="meta-body-item meta-body-info">
        <span class="date">3 septembre 2026</span>
        <span class="spacer">|</span>
        2h 17min
        <span class="spacer">|</span>
        <span class="dark-grey-link">Concert</span>,
        <span class="dark-grey-link">Musical</span>
      </div>
      <div class="meta-body-item meta-body-direction">
        <span class="light">De</span>
        <span class="dark-grey-link">Paul Dugdale</span>
      </div>
      <div class="meta-body-item meta-body-actor">
        <span class="light">Avec</span>
        <a class="dark-grey-link" href="/personne/fichepersonne_gen_cpersonne=424661.html">Katy Perry</a>
      </div>
      <div class="meta-body-item">
        <span class="light">Titre original </span>
        <span class="dark-grey">Katy Perry: The Lifetimes Tour</span>
      </div>
    </div>
    <div class="js-affinity-badge" data-entity-id="TW92aWU6MTAwMDA0NTYwOQ=="></div>
  </div>
</div>
<div class="rating-holder rating-holder-3">
  <div class="rating-item">
    <div class="rating-item-content">
      <span class="rating-title"> Presse </span>
      <div class="stareval stareval-small stareval-theme-default">
        <span class="stareval-note no-rating"></span>
      </div>
    </div>
  </div>
  <div class="rating-item">
    <div class="rating-item-content">
      <span class="rating-title"> Spectateurs </span>
      <div class="stareval stareval-small stareval-theme-default">
        <span class="stareval-note no-rating"></span>
      </div>
    </div>
  </div>
</div>
<div class="synopsis"><div class="content-txt">Concert à l'Accor Arena de Paris.</div></div>
</body></html>`

func TestParse(t *testing.T) {
	releases, err := Parse(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(releases) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(releases))
	}

	r1 := releases[0]
	if r1.ID != "TW92aWU6MzAwNzg0" {
		t.Errorf("ID = %q", r1.ID)
	}
	if r1.Title != "La Dernière patiente" {
		t.Errorf("Title = %q", r1.Title)
	}
	if r1.ReleaseDate != "2 septembre 2026" {
		t.Errorf("ReleaseDate = %q", r1.ReleaseDate)
	}
	if r1.Duration != "1h 20min" {
		t.Errorf("Duration = %q", r1.Duration)
	}
	if r1.DurationMinutes != 80 {
		t.Errorf("DurationMinutes = %d", r1.DurationMinutes)
	}
	if len(r1.Genres) != 1 || r1.Genres[0] != "Drame" {
		t.Errorf("Genres = %v", r1.Genres)
	}
	if r1.Director != "Rémi Bassaler" {
		t.Errorf("Director = %q", r1.Director)
	}
	if len(r1.Actors) != 2 || r1.Actors[0] != "Anouk Grinberg" || r1.Actors[1] != "Félix Lefebvre" {
		t.Errorf("Actors = %v", r1.Actors)
	}
	if r1.Poster != "https://fr.web.img2.acsta.net/c_310_420/img/abc.jpg" {
		t.Errorf("Poster = %q", r1.Poster)
	}
	if r1.Link != "https://www.allocine.fr/film/fichefilm_gen_cfilm=300784.html" {
		t.Errorf("Link = %q", r1.Link)
	}
	if r1.Synopsis != "Film sur une patiente." {
		t.Errorf("Synopsis = %q", r1.Synopsis)
	}
	if r1.PressRating != 2.8 {
		t.Errorf("PressRating = %f", r1.PressRating)
	}
	if r1.SpectatorRating != 3.4 {
		t.Errorf("SpectatorRating = %f", r1.SpectatorRating)
	}

	r2 := releases[1]
	if r2.Title != "Katy Perry: The Lifetimes Tour" {
		t.Errorf("r2.Title = %q", r2.Title)
	}
	if r2.OriginalTitle != "Katy Perry: The Lifetimes Tour" {
		t.Errorf("r2.OriginalTitle = %q", r2.OriginalTitle)
	}
	if len(r2.Genres) != 2 || r2.Genres[0] != "Concert" || r2.Genres[1] != "Musical" {
		t.Errorf("r2.Genres = %v", r2.Genres)
	}
	if r2.Duration != "2h 17min" {
		t.Errorf("r2.Duration = %q", r2.Duration)
	}
	if r2.DurationMinutes != 137 {
		t.Errorf("r2.DurationMinutes = %d", r2.DurationMinutes)
	}
	if r2.PressRating != 0 || r2.SpectatorRating != 0 {
		t.Errorf("r2 ratings should be 0, got press=%f spectator=%f", r2.PressRating, r2.SpectatorRating)
	}
}

func TestParseEmpty(t *testing.T) {
	releases, err := Parse(strings.NewReader("<html><body></body></html>"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(releases) != 0 {
		t.Fatalf("expected 0 releases, got %d", len(releases))
	}
}
