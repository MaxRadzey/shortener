// Package exitpkg — тестовый пакет для проверки диагностики вызова os.Exit вне main.
package exitpkg

import "os"

// F вызывает os.Exit(1) — анализатор должен сообщить о нарушении.
func F() {
	os.Exit(1) // want "os.Exit must not be used outside main.main"
}
