// Package panicpkg — тестовый пакет для проверки диагностики вызова panic.
package panicpkg

// F вызывает panic — анализатор должен сообщить о нарушении.
func F() {
	panic("x") // want "use of builtin panic is not allowed"
}
