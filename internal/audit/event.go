package audit

import "time"

// Event — событие аудита запроса.
type Event struct {
	TS     int64  `json:"ts"`      // unix timestamp события
	Action string `json:"action"`  // "shorten" (создание) или "follow" (прохождение по ссылке)
	UserID string `json:"user_id"` // идентификатор пользователя, если есть
	URL    string `json:"url"`     // оригинальный (не сокращённый) URL
}

// NewEvent создаёт событие с текущим временем.
func NewEvent(action, userID, url string) Event {
	return Event{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
