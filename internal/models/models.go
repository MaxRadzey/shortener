package models

// Request — тело запроса POST /api/shorten.
type Request struct {
	URL string `json:"url"`
}

// Response — ответ с короткой ссылкой.
type Response struct {
	Result string `json:"result"`
}

// BatchRequestItem — элемент запроса batch (correlation_id + original_url).
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponseItem — элемент ответа batch (correlation_id + short_url).
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLItem — элемент списка ссылок пользователя.
type UserURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
