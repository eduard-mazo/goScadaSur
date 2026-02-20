// web/embed.go
package web

import (
	"embed"
	"io/fs"
	"net/http"
)

// DistFS contiene los archivos estáticos de la UI
//go:embed dist/*
var DistFS embed.FS

// GetFS retorna el sistema de archivos para el servidor HTTP
func GetFS() http.FileSystem {
	sub, err := fs.Sub(DistFS, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
