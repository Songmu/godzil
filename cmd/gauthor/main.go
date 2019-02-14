package main

import (
	"flag"
	"log"
	"os"

	"github.com/Songmu/gauthor"
)

func main() {
	log.SetFlags(0)
	err := gauthor.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil && err != flag.ErrHelp {
		log.Println(err)
		os.Exit(1)
	}
}
