// Тесты анализатора через analysistest: запускается линтер на коде из testdata,
// а ожидаемые срабатывания задаются комментариями // want "текст" в конце строки.
package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer проверяет, что анализатор выдаёт нужные диагностики в тестовых пакетах.
//
// Как устроен тест:
//   - analysistest.Run запускает наш Analyzer для каждого переданного пакета из testdata/src/<имя>/.
//   - В исходниках тестов в конце строки пишется // want "ожидаемое сообщение".
//   - analysistest сверяет реальные диагностики анализатора с этими ожиданиями: количество и текст должны совпадать.
//
// Пакеты в testdata/src/:
//
//   - panicpkg    — вызов panic(); должна быть диагностика «use of builtin panic is not allowed».
//   - fatalpkg   — вызов log.Fatal() не в main; «log.Fatal/... must not be used outside main.main».
//   - exitpkg     — вызов os.Exit() не в main; «os.Exit must not be used outside main.main».
//   - mainok      — package main, в main() только os.Exit(0); диагностик быть не должно (разрешённый случай).
//   - mainpanic   — package main, в main() вызов panic(); диагностика про panic.
//   - ignorepkg   — panic() с комментарием // linter:ignore на той же строке; диагностик не должно быть (подавление).
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "panicpkg", "fatalpkg", "exitpkg", "mainok", "mainpanic", "ignorepkg")
}
