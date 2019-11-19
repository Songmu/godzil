package godzil

import (
	"io"

	"github.com/Songmu/gocredits"
)

type credits struct {
}

func (cr *credits) run(argv []string, outStream, errStream io.Writer) error {
	return gocredits.Run(argv, outStream, errStream)
}
