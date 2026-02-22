// Package audit — события аудита и рассылка по наблюдателям (файл, удалённый URL).
package audit

import "time"

// Event — одно событие аудита (время, действие shorten/follow, user_id, url).
type Event struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
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
