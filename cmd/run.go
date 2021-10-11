package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func (ctx *CmdCtx) printUsageAndExit() {
	ds := strings.Builder{}
	var i = 0
	for c := range ctx.Commands {
		if i > 0 {
			ds.WriteString(",")
		}
		ds.WriteString(c)
		i++
	}
	fmt.Fprintf(os.Stderr, "USAGE: %s [%v] [...]\n", os.Args[0], ds.String())
	os.Exit(1)
}

func cleanup() {
	s, err := unix.IoctlGetTermios(0, unix.TCGETS)
	if err != nil {
		return
	}
	if s.Lflag&unix.ECHO != 0 {
		return
	}
	s.Lflag |= unix.ECHO
	if err := unix.IoctlSetTermios(0, unix.TCSETS, s); err != nil {
		return
	}
}

func Run() {

	ctx, err := NewCtx()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Couldnt initialize, err was:\n%v\n",
			err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		ctx.printUsageAndExit()
	}

	c := ctx.Commands[os.Args[1]]
	if c == nil {
		fmt.Fprintf(
			os.Stderr,
			"Command '%s' not found, Use 'help' to list available commands\n",
			os.Args[1])
		os.Exit(1)
	}

	nc := make(chan os.Signal)
	signal.Notify(nc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-nc
		cleanup()
		os.Exit(0)
	}()

	if err := c.Handler(ctx, os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return
}
