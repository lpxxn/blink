package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

const (
	MaxImageBytes = 5 << 20 // 5 MiB
)

var allowedMimes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

var ErrInvalidImage = errors.New("upload: invalid image type or size")

// ValidateImageFile checks size and magic-bytes MIME of an uploaded file (first bytes only).
func ValidateImageFile(fh *multipart.FileHeader) error {
	if fh == nil {
		return ErrInvalidImage
	}
	if fh.Size <= 0 || fh.Size > MaxImageBytes {
		return fmt.Errorf("%w: size", ErrInvalidImage)
	}
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return err
	}
	if n == 0 {
		return ErrInvalidImage
	}
	mt := mimetype.Detect(buf[:n])
	if mt == nil {
		return ErrInvalidImage
	}
	m := mt.String()
	if _, ok := allowedMimes[m]; !ok {
		// some detectors return vendor-specific strings
		if !strings.HasPrefix(m, "image/") {
			return ErrInvalidImage
		}
	}
	return nil
}

// MIMEFromHeaderSniff returns MIME from file header bytes (for extension choice).
func MIMEFromHeaderSniff(fh *multipart.FileHeader) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}
	mt := mimetype.Detect(buf[:n])
	if mt == nil {
		return "", ErrInvalidImage
	}
	return mt.String(), nil
}

// ExtForMIME returns a file extension including dot, or empty.
func ExtForMIME(m string) string {
	switch m {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
