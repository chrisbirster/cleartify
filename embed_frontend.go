package cleartify

import (
	"embed"
	"io/fs"
)

//go:embed web/dist
var frontendFiles embed.FS

func Frontend() fs.FS {
	return subFrontend(frontendFiles, "web/dist")
}

func subFrontend(filesystem fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(filesystem, dir)
	if err != nil {
		panic(err)
	}

	return sub
}
