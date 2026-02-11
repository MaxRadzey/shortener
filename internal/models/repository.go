package models

// URLEntry — запись для создания URL в репозитории.
type URLEntry struct {
	ShortPath string
	FullURL   string
	UserID    string
	IsDeleted bool
}

// UserURL — short_path + original_url, используется в GetByUserID.
type UserURL struct {
	ShortPath   string
	OriginalURL string
}
