// Package migration embeds the goose SQL migrations so they ship inside the binary.
package migration

import "embed"

// FS holds every migration file in this directory.
//
//go:embed *.sql
var FS embed.FS
