// Package fatalpkg — тестовый пакет для проверки диагностики log.Fatal вне main.
package fatalpkg

import "log"

// F вызывает log.Fatal — анализатор должен сообщить о нарушении.
func F() {
	log.Fatal("err") // want "log.Fatal/Fatalf/Fatalln must not be used outside main.main"
}
