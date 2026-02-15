package audit

// Observer — приёмник событий аудита (файл, удалённый сервер и т.д.).
type Observer interface {
	Notify(event Event)
}

// Notifier рассылает события всем зарегистрированным наблюдателям.
type Notifier struct {
	observers []Observer
}

// NewNotifier создаёт нотифаер: при auditFile — пишет в файл, при auditURL — шлёт на URL.
func NewNotifier(auditFile, auditURL string) *Notifier {
	var observers []Observer
	if auditFile != "" {
		observers = append(observers, NewFileReceiver(auditFile))
	}
	if auditURL != "" {
		observers = append(observers, NewRemoteReceiver(auditURL))
	}
	return &Notifier{observers: observers}
}

// Notify рассылает событие всем наблюдателям.
func (n *Notifier) Notify(event Event) {
	for _, o := range n.observers {
		o.Notify(event)
	}
}
