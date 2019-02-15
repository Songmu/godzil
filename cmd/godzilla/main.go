package main

import (
	"flag"
	"log"
	"os"

	"github.com/Songmu/godzilla"
)

func main() {
	log.SetFlags(0)
	err := godzilla.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil && err != flag.ErrHelp {
		log.Println(err)
		os.Exit(1)
	}
}
