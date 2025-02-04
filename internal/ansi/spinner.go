package ansi

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

const (
	spinnerTextEllipsis = "..."
	spinnerTextDone     = "done"
	spinnerTextFailed   = "failed"

	spinnerColor = "red"
)

func Spinner(text string, fn func() error) error {
	done := make(chan struct{})
	errc := make(chan error)
	go func() {
		defer close(done)

		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		s.Prefix = text + spinnerTextEllipsis + " "
		s.FinalMSG = s.Prefix + spinnerTextDone

		if err := s.Color(spinnerColor); err != nil {
			panic(err)
		}

		s.Start()
		err := <-errc
		if err != nil {
			s.FinalMSG = s.Prefix + spinnerTextFailed
		}

		s.Stop()
	}()

	err := fn()
	errc <- err
	<-done
	fmt.Println()
	return err
}
