// Генерирует Reset() для структур с // generate:reset.
// Запуск:
//
//	go run ./cmd/reset ./...
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer)
}
