// Package ui embeds the gallery's browser front end so the binary serves it
// with no files on disk.
package ui

import (
	"embed"
	"io/fs"
)

//go:embed static
var files embed.FS

// Static is the asset tree served under /static/.
func Static() fs.FS {
	sub, err := fs.Sub(files, "static")
	if err != nil {
		panic(err) // the embed pattern above guarantees the directory exists
	}
	return sub
}

// IndexHTML is the gallery page.
func IndexHTML() []byte {
	b, err := files.ReadFile("static/index.html")
	if err != nil {
		panic(err)
	}
	return b
}
