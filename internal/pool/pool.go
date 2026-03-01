package pool

import "sync"

// Resettable — ограничение для типов, которые умеют сбрасывать своё состояние
// перед повторным использованием.
type Resettable interface {
	Reset()
}

// Pool — типобезопасная обёртка над sync.Pool для значений, реализующих Reset().
//
// Обычно T — указатель на структуру со сгенерированным методом Reset().
type Pool[T Resettable] struct {
	p sync.Pool
}

// New создаёт новый Pool, который создаёт значения через newFn, когда пул пуст.
// newFn не должен быть nil.
func New[T Resettable](newFn func() T) *Pool[T] {
	pl := &Pool[T]{}
	pl.p.New = func() any { return newFn() }
	return pl
}

// Get возвращает объект из пула (а при необходимости создаёт новый через New()).
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put сбрасывает состояние obj и возвращает его в пул.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.p.Put(obj)
}
