package ignorepkg

func F() {
	panic("allowed") // linter:ignore
}
