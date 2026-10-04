# zaphub

API JSON en Go pour le guide TV TNT française, les sorties cinéma de la semaine et les flux YouTube, avec une interface web embarquée.

Fournit :

- le **guide TV** (EPG) des chaînes TNT françaises à partir de [xmltvfr.fr](https://xmltvfr.fr),
- les **sorties cinéma** de la semaine via [AlloCiné](https://www.allocine.fr),
- les **dernières vidéos de chaînes YouTube** via les flux RSS/Atom officiels de chaque chaîne (sans clé API).

Les données sont téléchargées en arrière-plan, mises en cache en mémoire et exposées en JSON. L’application embarque également une interface web (SPA vanilla) accessible à la racine du serveur.

## Prérequis

- [mise](https://mise.jdx.dev) pour gérer les outils et les tâches.
- Sinon, [Go 1.26](https://go.dev) installé manuellement.
- Docker (optionnel) si tu veux exécuter l’image OCI publiée plutôt que compiler localement.

## Démarrage rapide

```sh
mise install
mise run bootstrap
go run ./cmd/zaphub
```

Le serveur écoute sur le port `8080` par défaut. Ouvre `http://localhost:8080` pour utiliser l’interface web.

Arrêt propre avec `Ctrl-C`.

## Commandes principales

| Commande | Description |
|---|---|
| `mise run bootstrap` | Installe les outils et les dépendances Go |
| `mise run build` | Compile les artefacts via GoReleaser (`./dist`) |
| `mise run lint` | Lance tous les linters (Markdown, JSON, Go) |
| `mise run fmt` | Formate les fichiers Go |
| `mise run verify` | Lance toutes les validations avant commit / PR |

Pour plus de détails sur les tâches disponibles, voir [`mise.toml`](mise.toml).

## Configuration

La configuration se fait par variables d’environnement (et `--log-level` en ligne de commande).

| Variable | Défaut | Description |
|---|---|---|
| `PORT` | `8080` | Port HTTP |
| `LOG_LEVEL` | `info` | Niveau de log : `debug`, `info`, `warn`, `error` |
| `EPG_URL` | `https://xmltvfr.fr/xmltv/xmltv_tnt.xml.gz` | Fichier XMLTV (`.xml.gz` recommandé, gzip natif) |
| `EPG_REFRESH_INTERVAL` | `6h` | Fréquence de rafraîchissement de l’EPG |
| `CINEMA_PROVIDER` | `allocine` | Fournisseur de sorties cinéma (`allocine`) |
| `CINEMA_URL` | `https://www.allocine.fr/film/sorties-semaine/` | URL source des sorties cinéma |
| `CINEMA_REFRESH_INTERVAL` | `24h` | Fréquence de rafraîchissement des sorties cinéma |
| `YT_REFRESH_INTERVAL` | `15m` | Fréquence de rafraîchissement des flux YouTube |
| `YT_CHANNELS_FILE` | `/data/youtube_channels.json` | Persistance de la liste des chaînes YouTube |
| `DATA_DIR` | `/data` | Répertoire de données |

## Interface web

`GET /` sert une interface web responsive sans dépendance externe :

- onglet **TV** : chaînes et programmes en cours / soirée,
- onglet **YouTube** : chaînes suivies et dernières vidéos,
- onglet **Cinéma** : sorties de la semaine.

La page `status.html` affiche l’état du service et des dernières mises à jour. Elle est accessible depuis la pastille de statut dans la barre supérieure.

## Endpoints JSON

### EPG

- `GET /healthz` — statut + dernier rafraîchissement de chaque source.
- `GET /api/epg/channels` — liste des chaînes `[{id, name, icon}]`.
- `GET /api/epg/now[?channel=TF1.fr]` — programme en cours sur chaque chaîne (ou une seule).
- `GET /api/epg/evening[?date=2026-08-12][&amp;channel=TF1.fr]` — les **2 programmes débutant à 21h00** pour chaque chaîne.
- `GET /api/epg/programmes?channel=TF1.fr[&amp;from=RFC3339][&amp;to=RFC3339]` — grille des programmes sur une plage (défaut : aujourd’hui).

### Cinéma

- `GET /api/cinema/releases` — sorties de la semaine (`[{id, title, original_title, poster, release_date, duration, duration_minutes, genres, director, actors, synopsis, press_rating, spectator_rating, link}]`).

### YouTube

- `GET /api/youtube/channels` — chaînes suivies.
- `POST /api/youtube/channels` — body `{"channel_id":"UC..."}` : ajoute une chaîne (résout le nom et charge les vidéos immédiatement).
- `DELETE /api/youtube/channels/{channel_id}` — supprime une chaîne.
- `GET /api/youtube/videos[?channel=UC...][&amp;limit=50]` — dernières vidéos, triées par date décroissante.

## Déploiement

Les releases produisent une image OCI légère publiée sur :

`ghcr.io/kazalab/zaphub:latest`

Exécution en local avec Docker :

```sh
docker run --rm -p 8080:8080 ghcr.io/kazalab/zaphub:latest
```

Pour persister la liste des chaînes YouTube, monte un volume sur `/data` :

```sh
docker run --rm -p 8080:8080 -v zaphub-data:/data ghcr.io/kazalab/zaphub:latest
```

## Consommation depuis Home Assistant

Exemples avec le `rest` sensor :

```yaml
sensor:
  - platform: rest
    name: "TV TF1"
    resource: "http://zaphub:8080/api/epg/now?channel=TF1.fr"
    scan_interval: 300
    value_template: "{{ value_json.programme.title if value_json.programme else 'aucun programme' }}"
    json_attributes_path: "$.programme"
    json_attributes:
      - description
      - start
      - stop
      - categories
      - rating

  - platform: rest
    name: "TV Soirée TF1"
    resource: "http://zaphub:8080/api/epg/evening?channel=TF1.fr"
    scan_interval: 3600
    value_template: "{{ value_json.channels[0].programmes[0].title }}"
    json_attributes_path: "$.channels[0]"
    json_attributes:
      - programmes

  - platform: rest
    name: "YouTube Dernière vidéo"
    resource: "http://zaphub:8080/api/youtube/videos?channel=UCXuqSBlHAE6Xw-yeJA0Tunw&amp;limit=1"
    scan_interval: 900
    value_template: "{{ value_json[0].title }}"
    json_attributes_path: "$[0]"
    json_attributes:
      - published
      - link
      - thumbnail
```

Ajout d'une chaîne YouTube à chaud :

```yaml
rest_command:
  zaphub_add_channel:
    url: "http://zaphub:8080/api/youtube/channels"
    method: POST
    content_type: "application/json"
    payload: '{"channel_id":"UCXuqSBlHAE6Xw-yeJA0Tunw"}'
```

## Version

```sh
zaphub version
```

Affiche la version, le commit et la date de build injectés par GoReleaser.

## Licence

[MIT](LICENSE)
