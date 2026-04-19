//go:build !embedweb

package bookaneer

import "embed"

// WebFS is a no-op placeholder used when the frontend has not been built.
// Build with -tags embedweb to include the real web/dist assets.
var WebFS embed.FS

// MigrationsFS contains the embedded SQL migration files.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
