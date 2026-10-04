// Package web embeds the web UI so the Go binary stays a single deployable
// artifact.
package web

import "embed"

// WebFS holds the web UI static assets.
//
//go:embed index.html app.js style.css status.html
var WebFS embed.FS
