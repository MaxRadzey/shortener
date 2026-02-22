package fatalpkg

import "log"

func F() {
	log.Fatal("err") // want "log.Fatal/Fatalf/Fatalln must not be used outside main.main"
}
