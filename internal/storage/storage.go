package storage

type UrlStorage interface {
	Get(short string) string
	Create(short, full string)
}

// Storage интерфейс хранения.
type Storage struct {
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{data: make(map[string]string)}
}

func (s *Storage) Get(short string) string {
	return s.data[short]
}

func (s *Storage) Create(short, full string) {
	s.data[short] = full
}
