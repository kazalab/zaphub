(() => {
  "use strict";

  const REFRESH_MS = 5 * 60 * 1000;
  const UNKNOWN_CHANNEL = "Chaîne inconnue";

  const $ = (sel) => document.querySelector(sel);
  const $$ = (sel) => Array.from(document.querySelectorAll(sel));

  const els = {
    tabs: $$(".tab"),
    tvNow: $("#tvNow"),
    tvEvening: $("#tvEvening"),
    tvNowCard: $("#tvNowCard"),
    tvEveningCard: $("#tvEveningCard"),
    tvSegBtns: $$(".seg-btn"),
    tvFilter: $("#tvFilter"),
    tvRefresh: $("#tvRefresh"),
    ytChannels: $("#ytChannels"),
    ytVideos: $("#ytVideos"),
    addForm: $("#addChannelForm"),
    channelIdInput: $("#channelIdInput"),
    addFeedback: $("#addFeedback"),
    cinema: $("#cinema"),
    cinemaRefresh: $("#cinemaRefresh"),
    health: $("#healthStatus"),
    statusPill: $("#statusPill"),
    statusDot: $("#statusDot"),
    statusText: $("#statusText"),
    toast: $("#toast"),
  };

  let tvFilterValue = "";
  let tvView = "now";

  // ---------- Utilities ----------

  function escapeHtml(str) {
    const div = document.createElement("div");
    div.textContent = str == null ? "" : String(str);
    return div.innerHTML;
  }

  function fmtTime(rfc3339) {
    const d = new Date(rfc3339);
    return d.toLocaleTimeString("fr-FR", { hour: "2-digit", minute: "2-digit" });
  }

  function fmtDateTime(rfc3339) {
    const d = new Date(rfc3339);
    return d.toLocaleString("fr-FR", {
      day: "2-digit",
      month: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function fmtDuration(pt) {
    if (!pt) return "";
    const m = pt.match(/PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?/);
    if (!m) return "";
    const h = m[1] ? parseInt(m[1], 10) : 0;
    const min = m[2] ? parseInt(m[2], 10) : 0;
    const s = m[3] ? parseInt(m[3], 10) : 0;
    if (h) return `${h}:${String(min).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
    return `${min}:${String(s).padStart(2, "0")}`;
  }

  function emptyState(target, msg) {
    target.innerHTML = `<div class="empty-state">${escapeHtml(msg)}</div>`;
  }

  function loading(target) {
    target.innerHTML = `<div class="empty-state"><span class="loader"></span></div>`;
  }

  function showToast(msg, type) {
    els.toast.textContent = msg;
    els.toast.className = "toast";
    if (type) els.toast.classList.add(type);
    els.toast.hidden = false;
    clearTimeout(showToast._t);
    showToast._t = setTimeout(() => { els.toast.hidden = true; }, 3500);
  }

  async function api(url, opts) {
    const res = await fetch(url, opts);
    const ct = res.headers.get("content-type") || "";
    const body = ct.includes("json") ? await res.json().catch(() => null) : null;
    if (!res.ok) {
      const msg = body && body.error ? body.error : `Erreur HTTP ${res.status}`;
      throw new Error(msg);
    }
    return body;
  }

  // ---------- Health / status ----------

  async function loadHealth() {
    try {
      const h = await api("/healthz");
      const epg = h.epg || {};
      const yt = h.youtube || {};
      const cinema = h.cinema || {};
      const err = epg.last_error || yt.last_error || cinema.last_error;
      els.statusDot.className = "status-dot " + (err ? "err" : "ok");
      els.statusText.textContent = err ? "Problème" : "En ligne";

      if (els.health) {
        els.health.innerHTML = [
          healthRow("Statut", h.status, h.status === "ok" ? "ok" : ""),
          healthRow("EPG · dernier refresh", epg.last_refresh ? fmtDateTime(epg.last_refresh) : "jamais"),
          healthRow("EPG · dernier succès", epg.last_success ? fmtDateTime(epg.last_success) : "jamais"),
          healthRow("EPG · dernière erreur", epg.last_error || "aucune", epg.last_error ? "err" : "ok"),
          healthRow("Cinéma · provider", cinema.provider || "—"),
          healthRow("Cinéma · sorties", String(cinema.releases ?? 0)),
          healthRow("Cinéma · dernier refresh", cinema.last_refresh ? fmtDateTime(cinema.last_refresh) : "jamais"),
          healthRow("Cinéma · dernière erreur", cinema.last_error || "aucune", cinema.last_error ? "err" : "ok"),
          healthRow("YouTube · nombre de chaînes", String(yt.channels ?? 0)),
          healthRow("YouTube · dernier refresh", yt.last_refresh ? fmtDateTime(yt.last_refresh) : "jamais"),
          healthRow("YouTube · dernière erreur", yt.last_error || "aucune", yt.last_error ? "err" : "ok"),
        ].join("");
      }
    } catch (e) {
      els.statusDot.className = "status-dot err";
      els.statusText.textContent = "Injoignable";
      if (els.health) {
        els.health.innerHTML = `<div class="empty-state">Impossible de contacter le service : ${escapeHtml(e.message)}</div>`;
      }
    }
  }

  function healthRow(label, value, cls) {
    const v = value ?? "—";
    return `<div class="health-item"><span class="health-label">${escapeHtml(label)}</span>
      <span class="health-value ${cls || ""}">${escapeHtml(v)}</span></div>`;
  }

  // ---------- TV Guide ----------

  function channelIcon(id) {
    return `https://www.tvlux.be/images/logos/${id.replace(".fr", "").toLowerCase()}.png`;
  }

  function filterChannel(name) {
    if (!tvFilterValue) return true;
    return name.toLowerCase().includes(tvFilterValue);
  }

  function progBadges(p) {
    if (!p) return "";
    const chips = [];
    (p.categories || []).slice(0, 3).forEach((c) => {
      chips.push(`<span class="badge">${escapeHtml(c)}</span>`);
    });
    if (p.rating) chips.push(`<span class="badge badge-rating">${escapeHtml(p.rating)}</span>`);
    return chips.length ? `<div class="prog-badges">${chips.join("")}</div>` : "";
  }

  function progPoster(p) {
    if (!p || !p.icon) return "";
    return `<img class="prog-poster" src="${escapeHtml(p.icon)}" alt="" loading="lazy" onerror="this.remove()">`;
  }

  function progExtras(p) {
    if (!p || !(p.sub_title || p.description)) return "";
    const sub = p.sub_title ? `<p class="prog-subtitle">${escapeHtml(p.sub_title)}</p>` : "";
    const desc = p.description ? `<p>${escapeHtml(p.description)}</p>` : "";
    return `<div class="prog-extras">
      <button type="button" class="details-toggle">Détails</button>
      <div class="prog-desc" hidden>${sub}${desc}</div>
    </div>`;
  }

  function wireDetails(container) {
    $$(".details-toggle", container).forEach((btn) => {
      btn.addEventListener("click", () => {
        const desc = btn.nextElementSibling;
        const open = !desc.hidden;
        desc.hidden = open;
        btn.textContent = open ? "Détails" : "Masquer";
      });
    });
  }

  async function loadNow() {
    const list = els.tvNow;
    loading(list);
    try {
      const entries = await api("/api/epg/now");
      const html = entries
        .filter((e) => filterChannel(e.channel_name || e.channel_id))
        .map((e) => {
          const p = e.programme;
          const icon = e.icon
            ? `<img class="channel-icon" src="${escapeHtml(e.icon)}" alt="" onerror="this.style.opacity=0">`
            : `<div class="channel-icon" style="display:flex;align-items:center;justify-content:center;color:#5b6472">TV</div>`;
          const prog = p
            ? `${progPoster(p)}
               <div class="now-title">${escapeHtml(p.title)}</div>
               <div class="now-time">${fmtTime(p.start)} – ${fmtTime(p.stop)}</div>
               ${progBadges(p)}
               ${progExtras(p)}`
            : `<div class="now-empty">Aucun programme</div>`;
          return `<div class="now-item">
            ${icon}
            <div class="channel-meta">
              <div class="channel-name">${escapeHtml(e.channel_name || e.channel_id)}</div>
            </div>
            <div class="now-prog">${prog}</div>
          </div>`;
        })
        .join("");
      list.innerHTML = html || `<div class="empty-state">Aucune chaîne ne correspond au filtre</div>`;
      wireDetails(list);
    } catch (e) {
      list.innerHTML = `<div class="empty-state">Erreur : ${escapeHtml(e.message)}</div>`;
    }
  }

  async function loadEvening() {
    const grid = els.tvEvening;
    loading(grid);
    try {
      const today = new Date().toISOString().slice(0, 10);
      const resp = await api(`/api/epg/evening?date=${today}`);
      const items = resp.channels.filter((c) => filterChannel(c.channel_name || c.channel_id));
      const html = items
        .map((c) => {
          const progs = c.programmes
            .map((p) => `<div class="evening-prog">
                ${progPoster(p)}
                <div class="evening-prog-time">${fmtTime(p.start)} – ${fmtTime(p.stop)}</div>
                <div class="evening-prog-title">${escapeHtml(p.title)}</div>
                ${progBadges(p)}
                ${progExtras(p)}
              </div>`)
            .join("");
          const icon = c.icon
            ? `<img src="${escapeHtml(c.icon)}" alt="" onerror="this.style.opacity=0">`
            : `<div style="width:32px;height:32px;border-radius:6px;background:#fff;"></div>`;
          return `<div class="evening-card">
            <div class="evening-head">${icon}<div class="evening-ch-name">${escapeHtml(c.channel_name || c.channel_id)}</div></div>
            ${progs}
          </div>`;
        })
        .join("");
      grid.innerHTML = html || `<div class="empty-state">Aucune chaîne ne correspond au filtre</div>`;
      wireDetails(grid);
    } catch (e) {
      grid.innerHTML = `<div class="empty-state">Erreur : ${escapeHtml(e.message)}</div>`;
    }
  }

  // ---------- Cinéma ----------

  async function loadCinema() {
    const grid = els.cinema;
    loading(grid);
    try {
      const releases = await api("/api/cinema/releases");
      const html = releases
        .map((r) => {
          const poster = r.poster
            ? `<img class="film-poster" src="${escapeHtml(r.poster)}" alt="" loading="lazy" onerror="this.style.display='none'">`
            : `<div class="film-poster-placeholder">🎬</div>`;
          const badges = [
            ...(r.genres || []).map((g) => `<span class="badge">${escapeHtml(g)}</span>`),
          ];
          if (r.press_rating) badges.push(`<span class="badge badge-rating">Presse ${r.press_rating.toFixed(1)}/5</span>`);
          if (r.spectator_rating) badges.push(`<span class="badge badge-rating">Spectateurs ${r.spectator_rating.toFixed(1)}/5</span>`);
          const extras = [
            r.director ? `<p><strong>Réalisateur :</strong> ${escapeHtml(r.director)}</p>` : "",
            r.actors && r.actors.length ? `<p><strong>Avec :</strong> ${escapeHtml(r.actors.join(", "))}</p>` : "",
            r.original_title ? `<p><strong>Titre original :</strong> ${escapeHtml(r.original_title)}</p>` : "",
            r.synopsis ? `<p>${escapeHtml(r.synopsis)}</p>` : "",
            r.link ? `<p><a href="${escapeHtml(r.link)}" target="_blank" rel="noopener">Voir sur AlloCiné →</a></p>` : "",
          ].filter(Boolean).join("");
          return `<div class="film-card">
            ${poster}
            <div class="film-info">
              <div class="film-title">${escapeHtml(r.title)}</div>
              <div class="film-meta">${escapeHtml(r.release_date || "")}${r.duration ? " · " + escapeHtml(r.duration) : ""}</div>
              <div class="prog-badges">${badges.join("")}</div>
              ${extras ? `<div class="prog-extras"><button type="button" class="details-toggle">Détails</button><div class="prog-desc" hidden>${extras}</div></div>` : ""}
            </div>
          </div>`;
        })
        .join("");
      grid.innerHTML = html || `<div class="empty-state">Aucune sortie pour le moment</div>`;
      wireDetails(grid);
    } catch (e) {
      grid.innerHTML = `<div class="empty-state">Erreur : ${escapeHtml(e.message)}</div>`;
    }
  }

  // ---------- YouTube ----------

  async function loadYTChannels() {
    const cont = els.ytChannels;
    loading(cont);
    try {
      const channels = await api("/api/youtube/channels");
      const html = channels
        .map((c) => {
          const name = c.name || UNKNOWN_CHANNEL;
          return `<div class="channel-row">
            <div class="chan-info">
              <div class="chan-name">${escapeHtml(name)}</div>
              <div class="chan-id">${escapeHtml(c.channel_id)}</div>
            </div>
            <button class="btn-danger" data-id="${escapeHtml(c.channel_id)}">Supprimer</button>
          </div>`;
        })
        .join("");
      cont.innerHTML = html || `<div class="empty-state">Aucune chaîne suivie</div>`;
      $$(".btn-danger[data-id]", cont).forEach((b) => {
        b.addEventListener("click", () => deleteYTChannel(b.dataset.id));
      });
    } catch (e) {
      cont.innerHTML = `<div class="empty-state">Erreur : ${escapeHtml(e.message)}</div>`;
    }
  }

  async function loadYTVideos() {
    const grid = els.ytVideos;
    loading(grid);
    try {
      const videos = await api("/api/youtube/videos?limit=50");
      const html = videos
        .map((v) => {
          const thumb = v.thumbnail
            ? `<div style="position:relative">
                 <img class="video-thumb" src="${escapeHtml(v.thumbnail)}" alt="" loading="lazy"
                      onerror="this.style.display='none'">
                 ${v.duration ? `<span class="video-duration">${escapeHtml(fmtDuration(v.duration))}</span>` : ""}
               </div>`
            : `<div class="video-thumb-placeholder">▶</div>`;
          const title = v.title || "Sans titre";
          const name = v.channel_name;
          return `<a class="video-card" href="${escapeHtml(v.link)}" target="_blank" rel="noopener noreferrer">
            ${thumb}
            <div class="video-body">
              <div class="video-title">${escapeHtml(title)}</div>
              <div class="video-meta">
                <span>${escapeHtml(name || v.channel_id || "")}</span>
                <span>${fmtDateTime(v.published)}</span>
              </div>
            </div>
          </a>`;
        })
        .join("");
      grid.innerHTML = html || `<div class="empty-state">Aucune vidéo pour le moment</div>`;
    } catch (e) {
      grid.innerHTML = `<div class="empty-state">Erreur : ${escapeHtml(e.message)}</div>`;
    }
  }

  async function addYTChannel(channelId) {
    els.addFeedback.hidden = true;
    try {
      await api("/api/youtube/channels", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ channel_id: channelId }),
      });
      els.addFeedback.textContent = "Chaîne ajoutée ✓";
      els.addFeedback.className = "feedback ok";
      els.addFeedback.hidden = false;
      await loadYTChannels();
      setTimeout(loadYTVideos, 1500);
    } catch (e) {
      els.addFeedback.textContent = `Erreur : ${e.message}`;
      els.addFeedback.className = "feedback err";
      els.addFeedback.hidden = false;
    }
  }

  async function deleteYTChannel(channelId) {
    try {
      await api(`/api/youtube/channels/${encodeURIComponent(channelId)}`, { method: "DELETE" });
      showToast(`Chaîne supprimée`, "ok");
      await Promise.all([loadYTChannels(), loadYTVideos()]);
    } catch (e) {
      showToast(`Erreur : ${e.message}`, "err");
    }
  }

  // ---------- Wiring ----------

  function refreshTVView() {
    if (tvView === "now") loadNow();
    else loadEvening();
  }

  function switchTVView(name) {
    tvView = name;
    els.tvSegBtns.forEach((b) => {
      const active = b.dataset.tvview === name;
      b.classList.toggle("active", active);
      b.setAttribute("aria-selected", String(active));
    });
    els.tvNowCard.hidden = name !== "now";
    els.tvEveningCard.hidden = name !== "evening";
    refreshTVView();
  }

  function switchTab(name) {
    els.tabs.forEach((t) => {
      const active = t.dataset.tab === name;
      t.classList.toggle("active", active);
      t.setAttribute("aria-selected", String(active));
    });
    const activePanel = $(`#tab-${name}`);
    $$(".tab-panel").forEach((p) => p.classList.remove("active"));
    activePanel.classList.add("active");

    if (name === "tv") {
      refreshTVView();
    } else if (name === "cinema") {
      loadCinema();
    } else if (name === "youtube") {
      loadYTChannels();
      loadYTVideos();
    }
  }

  // ---------- Init ----------

  const onStatusPage = document.body.hasAttribute("data-page-status");

  if (onStatusPage) {
    loadHealth();
    setInterval(loadHealth, 15000);
    return;
  }

  els.tabs.forEach((t) => {
    t.addEventListener("click", () => switchTab(t.dataset.tab));
  });

  els.tvSegBtns.forEach((b) => {
    b.addEventListener("click", () => switchTVView(b.dataset.tvview));
  });

  let debounce;
  els.tvFilter.addEventListener("input", () => {
    tvFilterValue = els.tvFilter.value.trim().toLowerCase();
    clearTimeout(debounce);
    debounce = setTimeout(refreshTVView, 200);
  });

  els.tvRefresh.addEventListener("click", () => {
    els.tvRefresh.classList.add("spin");
    setTimeout(() => els.tvRefresh.classList.remove("spin"), 500);
    refreshTVView();
  });

  els.cinemaRefresh.addEventListener("click", () => {
    els.cinemaRefresh.classList.add("spin");
    setTimeout(() => els.cinemaRefresh.classList.remove("spin"), 500);
    loadCinema();
  });

  els.addForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const id = els.channelIdInput.value.trim();
    if (!id) return;
    const btn = els.addForm.querySelector("button[type=submit]");
    btn.disabled = true;
    addYTChannel(id).finally(() => {
      btn.disabled = false;
      els.channelIdInput.value = "";
    });
  });

  loadHealth();
  refreshTVView();
  setInterval(() => {
    const active = $(".tab.active");
    const name = active ? active.dataset.tab : "tv";
    if (name !== "youtube") loadHealth();
    if (name === "tv") refreshTVView();
    else if (name === "cinema") loadCinema();
    else if (name === "youtube") loadYTVideos();
  }, REFRESH_MS);
})();
