package bookaneer

import "embed"

// WebFS contains the embedded frontend files.
//
//go:embed all:web/dist
var WebFS embed.FS

// MigrationsFS contains the embedded SQL migration files.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
