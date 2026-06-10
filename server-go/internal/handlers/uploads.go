package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/storage"
)

const maxAvatarBytes = 2 << 20 // 2 MB

// avatarExtensions maps detected content types to file extensions.
var avatarExtensions = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
}

// UploadHandler stores user-submitted files (currently avatars).
type UploadHandler struct {
	db    *sql.DB
	store storage.Store
}

func NewUploadHandler(db *sql.DB, store storage.Store) *UploadHandler {
	return &UploadHandler{db: db, store: store}
}

// UploadAvatar	godoc
// @Summary		Upload avatar
// @Description	Upload a profile image (png/jpeg/webp, max 2MB) as multipart field "file". Updates the user's avatarUrl.
// @Tags		profiles
// @Accept		mpfd
// @Produce		json
// @Security	BearerAuth
// @Param		file formData file true "Image file"
// @Success		200 {object} map[string]string
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Router		/profiles/me/avatar [post]
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A 'file' upload field is required"})
		return
	}
	if fileHeader.Size > maxAvatarBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image too large (max 2MB)"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read upload"})
		return
	}
	defer f.Close()

	// Sniff the real content type — never trust the client header.
	head := make([]byte, 512)
	n, _ := io.ReadFull(f, head)
	contentType := http.DetectContentType(head[:n])
	ext, allowed := avatarExtensions[strings.ToLower(contentType)]
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported image type (png, jpeg, or webp only)"})
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process upload"})
		return
	}

	key := fmt.Sprintf("avatars/%s.%s", userID, ext)
	url, err := h.store.Save(c.Request.Context(), key, io.LimitReader(f, maxAvatarBytes))
	if err != nil {
		slog.Error("avatar save failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store image"})
		return
	}

	if _, err := h.db.ExecContext(c.Request.Context(),
		`UPDATE users SET avatar_url = $1 WHERE id = $2`, url, userID,
	); err != nil {
		slog.Error("avatar url update failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatarUrl": url})
}
