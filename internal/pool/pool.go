package pool

import "sync"

// Resetter — тип с методом сброса состояния (например, сгенерированным через cmd/reset).
type Resetter interface {
	Reset()
}

// Pool — пул объектов типа T с методом Reset(), переиспользуемых через sync.Pool.
type Pool[T Resetter] struct {
	p sync.Pool
}

// New создаёт и возвращает указатель на пул для типа T.
// newFunc вызывается при необходимости создать новый объект, когда пул пуст.
func New[T Resetter](newFunc func() T) *Pool[T] {
	pool := &Pool[T]{}
	pool.p.New = func() any { return newFunc() }
	return pool
}

// Get возвращает объект из пула (или новый, если пул пуст).
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put сбрасывает состояние объекта и помещает его в пул.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.p.Put(x)
}
