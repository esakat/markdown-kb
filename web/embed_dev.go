//go:build dev

package web

import "embed"

// DistFS is an empty filesystem for development builds.
// Use `go build -tags dev` to build without frontend assets.
var DistFS embed.FS
