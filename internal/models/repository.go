package models

// URLEntry — одна запись URL в хранилище (short path, full URL, владелец, флаг удаления).
type URLEntry struct {
	ShortPath string
	FullURL   string
	UserID    string
	IsDeleted bool
}

// UserURL — пара short_path и original_url для выдачи по пользователю.
type UserURL struct {
	ShortPath   string
	OriginalURL string
}
