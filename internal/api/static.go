package api

import (
	"io/fs"
	"strings"

	"github.com/gin-gonic/gin"
)

var staticFS fs.FS

// SetStaticFS sets the filesystem for serving static files
func SetStaticFS(fsys fs.FS) {
	staticFS = fsys
}

func (s *Server) setupStatic() {
	if staticFS == nil {
		return
	}

	// Serve index.html at root
	s.router.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	// Serve assets
	s.router.GET("/assets/*filepath", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		data, err := fs.ReadFile(staticFS, path)
		if err != nil {
			c.Status(404)
			return
		}

		contentType := "application/octet-stream"
		if strings.HasSuffix(path, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(path, ".css") {
			contentType = "text/css"
		}
		c.Data(200, contentType, data)
	})

	// Serve other static files
	s.router.GET("/favicon.svg", func(c *gin.Context) {
		data, _ := fs.ReadFile(staticFS, "favicon.svg")
		c.Data(200, "image/svg+xml", data)
	})
	s.router.GET("/icons.svg", func(c *gin.Context) {
		data, _ := fs.ReadFile(staticFS, "icons.svg")
		c.Data(200, "image/svg+xml", data)
	})

	// SPA fallback
	s.router.NoRoute(func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})
}
