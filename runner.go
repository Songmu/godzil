package gauthor

import "io"

type runner interface {
	run([]string, io.Writer, io.Writer) error
}
