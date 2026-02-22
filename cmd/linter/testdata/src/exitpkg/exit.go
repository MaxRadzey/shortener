package exitpkg

import "os"

func F() {
	os.Exit(1) // want "os.Exit must not be used outside main.main"
}
