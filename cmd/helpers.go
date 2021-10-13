package cmd

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/kzaag/gnuflag"
	"golang.org/x/sys/unix"
)

type NameWithDesc struct {
	Name string
	Desc string
}

func GetHelpStr(commands []NameWithDesc) string {
	sb := strings.Builder{}
	for i := range commands {
		sb.WriteString(fmt.Sprintf("\t%s: %s\n", commands[i].Name, commands[i].Desc))
	}
	return sb.String()
}

// this is filled by the user
type CmdOptSpec struct {
	ArgFlags []string
	Label    string
	// for arguments with values.
	NotRequired    bool
	DefaultStrFunc func() string
	// password input
	NoEcho bool
	// flags without arguments
	IsBool bool
	// dont prompt for user input
	NoPrompt bool
}

// this is filled inside the function
type CmdOptVal struct {
	Str string
	// if argument was present within os.Args then set it to true (even if empty)
	Bool bool
}

type CmdOpt struct {
	Spec CmdOptSpec
	Val  CmdOptVal
}

func FilterArgs(allArgs []string) []string {
	ret := make([]string, 3)

	gnuflag.Getopt(allArgs, func(opt, optarg string) bool {
		if opt == "" {
			ret = append(ret, optarg)
		}
		return true
	})

	return ret
}

func GetOpts(args []string, reqOpts []CmdOpt) {

	flagHash := make(map[string]int)
	opts := make([]string, 0, len(reqOpts))

	for i := range reqOpts {
		for j := range reqOpts[i].Spec.ArgFlags {
			flag := reqOpts[i].Spec.ArgFlags[j]
			flagHash[flag] = i
			if reqOpts[i].Spec.IsBool {
				opts = append(opts, flag)
			} else {
				opts = append(opts, flag+":")
			}
		}
	}

	// skip name of program and previous argument
	gnuflag.Getopt(args, func(opt, optarg string) bool {

		// empty opt means this is an argument - ignore
		if opt == "" {
			return true
		}

		i, e := flagHash[opt]

		// if not found then print warning
		if !e || opt == "?" {
			fmt.Fprintf(os.Stderr, "%s: invalid option -- '%s'\n", os.Args[0], optarg)
			return true
		}

		reqOpts[i].Val.Str = optarg
		if reqOpts[i].Spec.IsBool {
			reqOpts[i].Val.Bool = true
		}
		return true
	}, opts...)

	// cli
	var tmp string
	fd := unix.Stdin
	reader := bufio.NewReader(os.Stdin)
	for i := 0; i < len(reqOpts); i++ {
		spec := &reqOpts[i].Spec
		tmp = ""

		// Argument was already set in os.Args
		if reqOpts[i].Val.Str != "" || (reqOpts[i].Val.Bool && spec.IsBool) {
			continue
		}

		// already set
		if reqOpts[i].Val.Str != "" {
			continue
		}

		if !spec.NoPrompt {
			// prompt for y/n
			if spec.IsBool {
				fmt.Print(spec.Label + " [y/n]: ")
				fmt.Scanln(&tmp)
				switch tmp {
				case "y":
					reqOpts[i].Val.Bool = true
				default:
					// dont set it
					continue
				}
				// prompt for value
			} else {
				fmt.Print(spec.Label + ": ")

				// read input without echo (eg. passwords)
				// note that you should handle sigint to reset echo before program exists
				if spec.NoEcho {
					s, err := unix.IoctlGetTermios(fd, unix.TCGETS)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Couldnt obtain terminal settings")
						os.Exit(1)
					}
					s.Lflag &^= unix.ECHO
					if err := unix.IoctlSetTermios(fd, unix.TCSETS, s); err != nil {
						fmt.Fprintln(os.Stderr, "Couldnt set terminal settings")
						os.Exit(1)
					}
					tmp, _ = reader.ReadString('\n')
					fmt.Println()
					s, err = unix.IoctlGetTermios(fd, unix.TCGETS)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Couldnt obtain terminal settings")
						os.Exit(1)
					}
					s.Lflag |= unix.ECHO
					if err := unix.IoctlSetTermios(fd, unix.TCSETS, s); err != nil {
						fmt.Fprintln(os.Stderr, "Couldnt set terminal settings")
						os.Exit(1)
					}
					// standard read input
				} else {
					tmp, _ = reader.ReadString('\n')
				}
			}

			if len(tmp) > 0 {
				tmp = tmp[:len(tmp)-1]
			}
			reqOpts[i].Val.Str = tmp

			// if still empty and was required - repeat
			if !spec.IsBool && !spec.NotRequired && tmp == "" {
				i--
				continue
			}
		}

		if tmp == "" && spec.DefaultStrFunc != nil {
			reqOpts[i].Val.Str = spec.DefaultStrFunc()
		}
	}
}

// dbg
// func (c CommandHandler) MarshalJSON() ([]byte, error) {
// 	if c == nil {
// 		return []byte("null"), nil
// 	} else {
// 		return []byte("\"" + runtime.FuncForPC(reflect.ValueOf(c).Pointer()).Name() + "\""), nil
// 	}
// }

type Branch struct {
	Str      string
	Command  *Command
	Branches []*Branch
}

func (b *Branch) FindOrInsertBranch(str string) *Branch {

	var mb *Branch
	for j := 0; j < len(b.Branches); j++ {
		if b.Branches[j].Str == str {
			mb = b.Branches[j]
			break
		}
	}

	if mb == nil {
		mb = &Branch{
			Str:      str,
			Command:  nil,
			Branches: make([]*Branch, 0, 2),
		}
		if b.Branches == nil {
			b.Branches = make([]*Branch, 0, 2)
		}
		b.Branches = append(b.Branches, mb)
	}

	return mb
}

func (b *Branch) SetCommandOrPanic(h *Command) {
	if b.Command != nil {
		panic("tried to add duplicate into tree")
	}
	b.Command = h
}

// o(n)
func (b *Branch) AddChainStrictOrder(h *Command, chain ...string) {
	var mb = b
	for i := 0; i < len(chain); i++ {
		mb = mb.FindOrInsertBranch(chain[i])
	}
	mb.SetCommandOrPanic(h)
}

func RemoveElementFromChain(chain []string, i int) []string {
	_c := make([]string, len(chain)-1)
	z := 0
	for j := range chain {
		if chain[j] != chain[i] {
			_c[z] = chain[j]
			z++
		}
	}
	return _c
}

func (b *Branch) AddChainAnyOrder(h *Command, chain ...string) []*Branch {
	var l = len(chain)
	var leaves = make([]*Branch, 0, 2)
	for i := 0; i < l; i++ {

		var mb = b.FindOrInsertBranch(chain[i])

		// this is leaf for this chain - add handler
		if l == 1 {
			mb.SetCommandOrPanic(h)
			return []*Branch{mb}
		}

		// create copy of c, but without current element
		chainWithoutOne := RemoveElementFromChain(chain, i)

		// repeat, but for child and without this element
		leaves = append(leaves, mb.AddChainAnyOrder(h, chainWithoutOne...)...)
	}

	return leaves
}

func (b *Branch) FindInChain(args []string) *Command {
	for i := range args {
		a := args[i]
		for j := range b.Branches {
			_b := b.Branches[j]
			if _b.Str != a {
				continue
			}
			if i == len(args)-1 {
				return _b.Command
			} else {
				return _b.FindInChain(RemoveElementFromChain(args, i))
			}
		}
	}
	return nil
}

const charset = "qwertyuiopasdfghjklzxcvbnm"

var maxInt = big.NewInt(100000)

// may panic
func randStr(l int) string {
	res := make([]byte, l)
	for i := 0; i < l; i++ {
		x, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			panic(err)
		}
		res[i] = charset[x.Int64()%int64(len(charset))]
	}
	return string(res)
}
