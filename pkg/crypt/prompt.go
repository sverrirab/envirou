package crypt

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"golang.org/x/term"
)

// ErrNoTerminal is returned when a passphrase is needed but no terminal is
// available to prompt on (pipes, CI).
var ErrNoTerminal = errors.New("cannot prompt for passphrase: no terminal available (run 'ev unlock' in an interactive shell or set ENVIROU_KEY)")

// ReadPassphrase prompts for a hidden passphrase. It is a variable so tests
// can stub it out.
var ReadPassphrase = readPassphraseTTY

func ttyDevice() string {
	if runtime.GOOS == "windows" {
		return "CONIN$"
	}
	return "/dev/tty"
}

// readPassphraseTTY reads from the controlling terminal so prompting works
// under command substitution (the ev wrapper) and with redirected stdin.
// Falls back to stdin only when stdin is a terminal.
func readPassphraseTTY(prompt string) (string, error) {
	if tty, err := os.OpenFile(ttyDevice(), os.O_RDWR, 0); err == nil {
		defer tty.Close()
		return readHidden(tty, tty, prompt)
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return readHidden(os.Stdin, os.Stderr, prompt)
	}
	return "", ErrNoTerminal
}

// readHidden prompts on w and reads a hidden line from in, restoring the
// terminal state if interrupted mid-read (ReadPassword disables echo).
func readHidden(in, w *os.File, prompt string) (string, error) {
	fd := int(in.Fd())
	state, _ := term.GetState(fd)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	done := make(chan struct{})
	go func() {
		select {
		case <-sigCh:
			if state != nil {
				_ = term.Restore(fd, state)
			}
			fmt.Fprintln(w)
			os.Exit(130)
		case <-done:
		}
	}()
	defer func() {
		close(done)
		signal.Stop(sigCh)
	}()

	fmt.Fprint(w, prompt)
	passphrase, err := term.ReadPassword(fd)
	fmt.Fprintln(w)
	if err != nil {
		return "", err
	}
	return string(passphrase), nil
}
