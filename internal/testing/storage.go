package testing

import (
	"context"

	"github.com/MaxRadzey/shortener/internal/storage"
)

// FakeStorage - мок хранилища для тестов.
type FakeStorage struct {
	data map[string]storage.URLEntry
}

// NewFakeStorage создает новый экземпляр FakeStorage с пустыми данными.
func NewFakeStorage() *FakeStorage {
	return &FakeStorage{
		data: make(map[string]storage.URLEntry),
	}
}

// NewFakeStorageWithData создает новый экземпляр FakeStorage с указанными данными.
// data - map[shortPath]fullURL
func NewFakeStorageWithData(data map[string]string) *FakeStorage {
	entries := make(map[string]storage.URLEntry)
	for short, fullURL := range data {
		entries[short] = storage.URLEntry{ShortPath: short, FullURL: fullURL, UserID: ""}
	}
	return &FakeStorage{data: entries}
}

// NewFakeStorageWithEntries создает новый экземпляр FakeStorage с указанными записями.
func NewFakeStorageWithEntries(entries map[string]storage.URLEntry) *FakeStorage {
	return &FakeStorage{data: entries}
}

func (f *FakeStorage) Get(short string) (string, error) {
	val, ok := f.data[short]
	if !ok {
		return "", &storage.ErrNotFound{ShortPath: short}
	}
	return val.FullURL, nil
}

func (f *FakeStorage) Create(item storage.URLEntry) error {
	f.data[item.ShortPath] = item
	return nil
}

func (f *FakeStorage) CreateBatch(ctx context.Context, items []storage.URLEntry) error {
	for _, item := range items {
		f.data[item.ShortPath] = item
	}
	return nil
}

func (f *FakeStorage) GetByUserID(ctx context.Context, userID string) ([]storage.UserURL, error) {
	var out []storage.UserURL
	for short, entry := range f.data {
		if entry.UserID == userID {
			out = append(out, storage.UserURL{ShortPath: short, OriginalURL: entry.FullURL})
		}
	}
	return out, nil
}

func (f *FakeStorage) Ping(ctx context.Context) error {
	return nil
}
