package webfs

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var distFS embed.FS

// GetFS returns the embedded web/dist filesystem
func GetFS() (fs.FS, error) {
	return fs.Sub(distFS, "web/dist")
}
