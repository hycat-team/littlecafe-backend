// Package migrations embeds this service's Goose migration files.
package migrations

import "embed"

// EmbedFS holds every *.sql migration of the service.
//
//go:embed *.sql
var EmbedFS embed.FS
