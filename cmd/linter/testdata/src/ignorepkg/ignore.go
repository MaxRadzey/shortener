// Package ignorepkg — тестовый пакет для проверки подавления диагностики через // linter:ignore.
package ignorepkg

// F вызывает panic с комментарием linter:ignore — диагностика не должна выводиться.
func F() {
	panic("allowed") // linter:ignore
}
