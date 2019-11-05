package godzil

import (
	"context"
	"io"

	"github.com/Songmu/ghch"
)

type ghchCmd struct {
}

func (gh *ghchCmd) run(argv []string, outStream, errStream io.Writer) error {
	return ghch.Run(context.Background(), argv, outStream, errStream)
}
