//go:build dev

package main

import "github.com/gin-gonic/gin"

// setupStatic is a no-op in dev mode.
// The Vite dev server on :4200 handles the frontend with hot reload.
func setupStatic(r *gin.Engine) {}
