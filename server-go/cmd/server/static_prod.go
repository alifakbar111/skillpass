//go:build !dev

package main

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/static"
)

// setupStatic mounts the embedded React SPA build into the Gin router.
// It serves index.html for all non-API, non-file routes (SPA catch-all).
func setupStatic(r *gin.Engine) {
	// Disable Gin's automatic redirect behaviour so SPA routes don't 301.
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	distFS, err := fs.Sub(static.FS, "dist")
	if err != nil {
		log.Printf("Warning: could not access embedded dist: %v", err)
		return
	}

	httpFS := http.FS(distFS)

	serveIndex := func(c *gin.Context) {
		// Read directly from the embedded FS to avoid http.FileServer's
		// automatic redirect of /index.html → ./
		data, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}

	isExcluded := func(path string) bool {
		return strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/uploads/") ||
			strings.HasPrefix(path, "/docs/") ||
			strings.HasPrefix(path, "/p/")
	}

	// Root — explicit route avoids 301 redirect.
	r.GET("/", func(c *gin.Context) {
		serveIndex(c)
	})

	// Serve hashed assets (JS, CSS, images) with immutable cache.
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.FileFromFS("/assets/"+c.Param("filepath"), httpFS)
	})

	// Other static files from dist root.
	r.GET("/favicon.ico", func(c *gin.Context) { c.FileFromFS("/favicon.ico", httpFS) })
	r.GET("/robots.txt", func(c *gin.Context) { c.FileFromFS("/robots.txt", httpFS) })

	// SPA catch-all: serve index.html for all non-API, non-file routes.
	r.NoRoute(func(c *gin.Context) {
		if isExcluded(c.Request.URL.Path) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		serveIndex(c)
	})

	log.Println("Serving embedded static files (production mode)")
}
