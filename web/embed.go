package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embeddedFS embed.FS

// GetFS returns the fs.FS of the built static assets.
func GetFS() (fs.FS, error) {
	return fs.Sub(embeddedFS, "dist")
}
