package godzil

import (
	"context"
	"io"

	"github.com/Songmu/goxz"
)

type crossbuild struct {
}

func (cb *crossbuild) run(argv []string, outStream, errStream io.Writer) error {
	return goxz.Run(context.Background(), argv, outStream, errStream)
}
