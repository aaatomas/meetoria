package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
)

const maxAvatarSize = 5 << 20 // 5MB

var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

func (s *LocalStorage) SaveEmployeeAvatar(orgID, employeeID string, file *multipart.FileHeader) (string, error) {
	if file.Size > maxAvatarSize {
		return "", apperrors.Validation("avatar must be 5MB or smaller")
	}

	src, err := file.Open()
	if err != nil {
		return "", apperrors.Internal("failed to read avatar", err)
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", apperrors.Internal("failed to read avatar", err)
	}

	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = strings.ToLower(detectContentType(buffer[:n]))
	}

	ext, ok := allowedAvatarTypes[contentType]
	if !ok {
		return "", apperrors.Validation("avatar must be a JPEG, PNG, or WebP image")
	}

	dir := filepath.Join(s.baseDir, "avatars", orgID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", apperrors.Internal("failed to prepare avatar directory", err)
	}

	filename := employeeID + ext
	destPath := filepath.Join(dir, filename)

	if err := s.removeExistingAvatar(dir, employeeID); err != nil {
		return "", err
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return "", apperrors.Internal("failed to save avatar", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, io.MultiReader(bytes.NewReader(buffer[:n]), src)); err != nil {
		return "", apperrors.Internal("failed to save avatar", err)
	}

	return fmt.Sprintf("/uploads/avatars/%s/%s", orgID, filename), nil
}

func (s *LocalStorage) DeleteByURL(url string) error {
	if url == "" || !strings.HasPrefix(url, "/uploads/") {
		return nil
	}
	path := filepath.Join(s.baseDir, strings.TrimPrefix(url, "/uploads/"))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) removeExistingAvatar(dir, employeeID string) error {
	for _, ext := range allowedAvatarTypes {
		path := filepath.Join(dir, employeeID+ext)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return apperrors.Internal("failed to replace avatar", err)
		}
	}
	return nil
}

func detectContentType(header []byte) string {
	if len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "image/jpeg"
	}
	if len(header) >= 8 && string(header[0:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP" {
		return "image/webp"
	}
	return ""
}
