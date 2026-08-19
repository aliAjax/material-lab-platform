package attachment

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestUploadChecksRealMIMEAndDownloadAuthorization(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(storage, 1024)
	if _, err = service.Upload(context.Background(), "commission", "id", "fake.pdf", "user", bytes.NewBufferString("plain text")); err == nil {
		t.Fatal("extension mismatch accepted")
	}
	object, err := service.Upload(context.Background(), "commission", "id", "note.txt", "user", bytes.NewBufferString("plain text"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Download(context.Background(), object, "other", func(string, string, string) bool { return false }); err == nil {
		t.Fatal("unauthorized download accepted")
	}
	reader, err := service.Download(context.Background(), object, "user", func(string, string, string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	content, _ := io.ReadAll(reader)
	if string(content) != "plain text" {
		t.Fatalf("content = %q", content)
	}
}

func TestLocalStorageRejectsTraversal(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = storage.Put(context.Background(), "../escape", bytes.NewBufferString("bad")); err == nil {
		t.Fatal("path traversal accepted")
	}
}
