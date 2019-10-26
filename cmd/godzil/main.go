package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/Songmu/godzil"
)

func main() {
	log.SetFlags(0)
	err := godzil.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
	if err != nil && err != flag.ErrHelp {
		log.Println(err)
		os.Exit(1)
	}
}
