package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStoreSave(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("UPLOAD_DIR", dir)
	defer os.Unsetenv("UPLOAD_DIR")

	store := NewLocalStore()
	if store.Dir() != dir {
		t.Fatalf("expected dir %q, got %q", dir, store.Dir())
	}

	url, err := store.Save(context.Background(), "avatars/user1.png", strings.NewReader("fake-image-bytes"))
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if url != "/uploads/avatars/user1.png" {
		t.Fatalf("unexpected url %q", url)
	}

	content, err := os.ReadFile(filepath.Join(dir, "avatars", "user1.png"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(content) != "fake-image-bytes" {
		t.Fatal("content mismatch")
	}
}

func TestLocalStoreRejectsTraversal(t *testing.T) {
	os.Setenv("UPLOAD_DIR", t.TempDir())
	defer os.Unsetenv("UPLOAD_DIR")
	store := NewLocalStore()

	for _, key := range []string{"../escape.txt", "/abs.txt", "a/../../b.txt", "bad key.txt"} {
		if _, err := store.Save(context.Background(), key, strings.NewReader("x")); err == nil {
			t.Fatalf("expected rejection for key %q", key)
		}
	}
}

func TestNewStoreFactory(t *testing.T) {
	if _, ok := NewStore().(*LocalStore); !ok {
		t.Fatal("expected LocalStore by default")
	}
}
