package attachment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"material-lab-platform/internal/domain"
)

type Object struct {
	ID          string    `json:"id"`
	OwnerType   string    `json:"ownerType"`
	OwnerID     string    `json:"ownerId"`
	StorageKey  string    `json:"-"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	SHA256      string    `json:"sha256"`
	UploadedBy  string    `json:"uploadedBy"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Storage interface {
	Put(context.Context, string, io.Reader) error
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) (*LocalStorage, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absolute, 0o750); err != nil {
		return nil, err
	}
	return &LocalStorage{root: absolute}, nil
}

func (s *LocalStorage) resolve(key string) (string, error) {
	if key == "" || filepath.Base(key) != key || strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid storage key")
	}
	path := filepath.Join(s.root, key)
	if !strings.HasPrefix(path, s.root+string(os.PathSeparator)) {
		return "", fmt.Errorf("storage path escaped root")
	}
	return path, nil
}

func (s *LocalStorage) Put(ctx context.Context, key string, source io.Reader) error {
	path, err := s.resolve(key)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(s.root, ".upload-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err = io.Copy(temporary, source); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}

func (s *LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.resolve(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := s.resolve(key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

type Service struct {
	storage Storage
	maxSize int64
	allowed map[string]bool
}

func NewService(storage Storage, maxSize int64) *Service {
	return &Service{storage: storage, maxSize: maxSize, allowed: map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/png":       true,
		"text/plain":      true,
	}}
}

func (s *Service) Upload(ctx context.Context, ownerType, ownerID, filename, actor string, source io.Reader) (Object, error) {
	if err := ctx.Err(); err != nil {
		return Object{}, err
	}
	if err := domain.Require(ownerType, "ownerType"); err != nil {
		return Object{}, err
	}
	if err := domain.Require(ownerID, "ownerId"); err != nil {
		return Object{}, err
	}
	if err := domain.Require(actor, "actor"); err != nil {
		return Object{}, err
	}
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "." || filename == "" {
		return Object{}, fmt.Errorf("%w: filename", domain.ErrValidation)
	}
	limited := io.LimitReader(source, s.maxSize+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return Object{}, err
	}
	if int64(len(content)) > s.maxSize {
		return Object{}, fmt.Errorf("%w: attachment too large", domain.ErrValidation)
	}
	if len(content) == 0 {
		return Object{}, fmt.Errorf("%w: empty attachment", domain.ErrValidation)
	}
	contentType := http.DetectContentType(content)
	if separator := strings.IndexByte(contentType, ';'); separator >= 0 {
		contentType = contentType[:separator]
	}
	if !s.allowed[contentType] {
		return Object{}, fmt.Errorf("%w: file type %s", domain.ErrValidation, contentType)
	}
	if !extensionMatches(filename, contentType) {
		return Object{}, fmt.Errorf("%w: extension does not match content", domain.ErrValidation)
	}
	key, err := domain.RandomToken(24)
	if err != nil {
		return Object{}, err
	}
	key += safeExtension(filename)
	if err = s.storage.Put(ctx, key, bytes.NewReader(content)); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		cleanupErr := s.storage.Delete(cleanupCtx, key)
		if cleanupErr != nil && !errors.Is(cleanupErr, os.ErrNotExist) {
			return Object{}, errors.Join(fmt.Errorf("store attachment: %w", err), fmt.Errorf("cleanup attachment: %w", cleanupErr))
		}
		return Object{}, fmt.Errorf("store attachment: %w", err)
	}
	digest := sha256.Sum256(content)
	return Object{ID: domain.NewID(), OwnerType: ownerType, OwnerID: ownerID, StorageKey: key, Filename: filename, ContentType: contentType, Size: int64(len(content)), SHA256: hex.EncodeToString(digest[:]), UploadedBy: actor, CreatedAt: time.Now().UTC()}, nil
}

func (s *Service) Download(ctx context.Context, object Object, actor string, authorized func(string, string, string) bool) (io.ReadCloser, error) {
	if authorized == nil || !authorized(actor, object.OwnerType, object.OwnerID) {
		return nil, domain.ErrForbidden
	}
	reader, err := s.storage.Open(ctx, object.StorageKey)
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrNotFound
	}
	return reader, err
}

func extensionMatches(filename, contentType string) bool {
	extension := strings.ToLower(filepath.Ext(filename))
	switch contentType {
	case "application/pdf":
		return extension == ".pdf"
	case "image/jpeg":
		return extension == ".jpg" || extension == ".jpeg"
	case "image/png":
		return extension == ".png"
	case "text/plain":
		return extension == ".txt"
	default:
		return false
	}
}

func safeExtension(filename string) string {
	extension := strings.ToLower(filepath.Ext(filename))
	if len(extension) > 8 {
		return ""
	}
	return extension
}
