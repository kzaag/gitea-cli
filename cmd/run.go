package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

// func (ctx *CmdCtx) printUsageAndExit() {
// 	ds := strings.Builder{}
// 	var i = 0
// 	for c := range ctx.Commands {
// 		if i > 0 {
// 			ds.WriteString(",")
// 		}
// 		ds.WriteString(c)
// 		i++
// 	}
// 	fmt.Fprintf(os.Stderr, "USAGE: %s [%v] [help,...]\n", os.Args[0], ds.String())
// 	os.Exit(1)
// }

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

	_ = ctx

	nc := make(chan os.Signal)
	signal.Notify(nc, os.Interrupt, syscall.SIGINT)
	go func() {
		<-nc
		cleanup()
		os.Exit(0)
	}()

	args := FilterArgs(os.Args[1:])

	c := ctx.CommandRoot.FindInChain(args)
	if c == nil {
		ctx.PrintCommands()
		os.Exit(1)
	}

	if err := c.Handler(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return
}
