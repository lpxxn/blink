package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	appupload "github.com/lpxxn/blink/application/upload"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) UploadImage(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.UploadRoot == "" || s.UploadURLPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "uploads not configured"})
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}
	if err := appupload.ValidateImageFile(fh); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mime, err := appupload.MIMEFromHeaderSniff(fh)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ext := appupload.ExtForMIME(mime)
	if ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type"})
		return
	}
	userDir := filepath.Join(s.UploadRoot, strconv.FormatInt(uid, 10))
	if err := os.MkdirAll(userDir, 0750); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	name, err := randomName()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filename := name + ext
	full := filepath.Join(userDir, filename)
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()
	dst, err := os.Create(full)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer dst.Close()
	if _, err := dst.ReadFrom(src); err != nil {
		_ = os.Remove(full)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	relURL := fmt.Sprintf("%s/%d/%s", strings.TrimRight(s.UploadURLPath, "/"), uid, filename)
	c.JSON(http.StatusOK, gin.H{"url": relURL})
}

func randomName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
