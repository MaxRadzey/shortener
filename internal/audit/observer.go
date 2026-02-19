package audit

import "sync"

// Observer — приёмник событий аудита (файл, удалённый сервер и т.д.).
type Observer interface {
	Notify(event Event)
}

// Notifier рассылает события всем зарегистрированным наблюдателям.
type Notifier struct {
	mu        sync.RWMutex
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

// Add добавляет наблюдателя. Потокобезопасно.
func (n *Notifier) Add(o Observer) {
	if o == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.observers = append(n.observers, o)
}

// Remove удаляет наблюдателя (по равенству интерфейса). Потокобезопасно.
func (n *Notifier) Remove(o Observer) {
	if o == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, obs := range n.observers {
		if obs == o {
			n.observers = append(n.observers[:i], n.observers[i+1:]...)
			return
		}
	}
}

// Notify рассылает событие всем наблюдателям.
func (n *Notifier) Notify(event Event) {
	n.mu.RLock()
	snapshot := make([]Observer, len(n.observers))
	copy(snapshot, n.observers)
	n.mu.RUnlock()

	for _, o := range snapshot {
		o.Notify(event)
	}
}
