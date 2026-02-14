package audit

// Observer — приёмник событий аудита.
type Observer interface {
	Notify(event Event)
}

// Notifier рассылает события всем зарегистрированным наблюдателям.
type Notifier struct {
	observers []Observer
}

// NewNotifier создаёт нотифаер с наблюдателями по конфигу.
// Если AuditFile задан — добавляется FileReceiver, если AuditURL — RemoteReceiver.
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

// Notify отправляет событие во все приёмники.
func (n *Notifier) Notify(event Event) {
	for _, o := range n.observers {
		o.Notify(event)
	}
}
