package godzil

import (
	"context"
	"io"

	"github.com/Songmu/ghch"
)

type changelog struct {
}

func (gh *changelog) run(argv []string, outStream, errStream io.Writer) error {
	return ghch.Run(context.Background(), argv, outStream, errStream)
}
