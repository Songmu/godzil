package {{.Package}}

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

const cmdName = "{{.Package}}"

// Run the {{.Package}}
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	ver := fs.Bool("version", false, "display version")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	eg, ctx := errgroup.WithContext(ctx)

	srv := &http.Server{Addr: ":" + port, Handler: http.HandlerFunc(index)}
	eg.Go(func() error {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("received signal %s, shutting down\n", sig)
	eg.Go(func() error {
		return srv.Shutdown(ctx)
	})
	return eg.Wait()

	return nil
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}
