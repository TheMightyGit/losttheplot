package main

import (
	"log"

	"github.com/TheMightyGit/losttheplot/cartridge"
	"github.com/TheMightyGit/marv/marvlib"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	marvlib.API.ConsoleBoot(
		"losttheplot",
		cartridge.Resources,
		cartridge.Start,
		cartridge.Update,
	)
}
