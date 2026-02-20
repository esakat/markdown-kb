//go:build !dev

package web

import "embed"

// DistFS embeds the built frontend assets from web/dist.
// Build with `make frontend` before `go build` to populate dist/.
//
//go:embed all:dist
var DistFS embed.FS
