// Тестовый main-пакет с panic в main() — ожидается диагностика про panic.
package main

func main() {
	panic("x") // want "use of builtin panic is not allowed"
}
