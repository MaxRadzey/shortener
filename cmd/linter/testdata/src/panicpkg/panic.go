package panicpkg

func F() {
	panic("x") // want "use of builtin panic is not allowed"
}
