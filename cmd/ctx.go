package cmd

import (
	"fmt"
	"gitea-cli/config"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type CommandHandler func() error

type Command struct {
	Desc    string
	Handler CommandHandler
	Opts    []CmdOpt
}

type CmdCtx struct {
	CommandRoot Branch
	Config      *config.Config
}

type commandPathInfo struct {
	Path    string
	Command *Command
}

func getCommands(b *Branch, parent string, commands *[]commandPathInfo) {
	for i := 0; i < len(b.Branches); i++ {
		_b := b.Branches[i]
		path := fmt.Sprintf("%s %s", parent, _b.Str)
		if _b.Command != nil {
			//fmt.Println(path)
			*commands = append(*commands, commandPathInfo{
				Path:    path,
				Command: _b.Command,
			})
		}
		getCommands(_b, path, commands)
	}
}

func (ctx *CmdCtx) PrintCommands() {
	commands := make([]commandPathInfo, 0, 4)
	getCommands(&ctx.CommandRoot, os.Args[0], &commands)

	// group commands by handler

	type groupedCommand struct {
		Command *Command
		Paths   []string
	}

	gc := make(map[string]groupedCommand)

	for i := range commands {
		c := commands[i]
		g, e := gc[c.Command.Desc]
		if !e {
			g = groupedCommand{
				Command: c.Command,
				Paths:   make([]string, 0, 2),
			}
		}
		g.Paths = append(g.Paths, c.Path)
		gc[c.Command.Desc] = g
	}

	fmt.Print("Usage: \n\n")

	for c := range gc {
		fmt.Println(c)
		for i := range gc[c].Paths {
			fmt.Printf("\t%s\n", gc[c].Paths[i])
		}

		if gc[c].Command.Opts != nil {

			for i := range gc[c].Command.Opts {
				o := gc[c].Command.Opts[i]
				if len(o.Spec.ArgFlags) == 0 {
					continue
				}
				if i == 0 {
					fmt.Println("Arguments:")
				}
				for j := range o.Spec.ArgFlags {
					os := o.Spec.ArgFlags[j]
					if j == 0 {
						fmt.Print("\t")
					}
					if j > 0 {
						fmt.Print(", ")
					}
					if len(os) > 1 {
						fmt.Printf("--%s", os)
					} else {
						fmt.Printf("-%s", os)
					}
				}
				fmt.Printf("\t%s\n", o.Spec.Label)
			}
		}
		fmt.Print("\n\n")
	}
}

// create context using saved config
func NewCtx() (*CmdCtx, error) {
	ctx := new(CmdCtx)

	fc, err := ioutil.ReadFile("gitea.yml")
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		cnf := new(config.Config)
		if err := yaml.Unmarshal(fc, cnf); err != nil {
			return nil, err
		}
		if err := cnf.Validate(); err != nil {
			fmt.Println("Your config is out of date. You should create it again")
			fmt.Print("Continue anyway? y/n ")
			var x string
			fmt.Scanln(&x)
			if x != "y" {
				return nil, err
			}
		}
		fmt.Printf("Using config for user %s\n", cnf.Username)
		ctx.Config = cnf
	}

	root := Branch{}

	root.AddChainAnyOrder(&Command{
		Desc:    "Help",
		Handler: ctx.HelpCommand,
	}, "help")
	root.AddChainAnyOrder(&Command{
		Desc:    "Delete local config with its token.",
		Handler: ctx.RmConfigCommand,
		Opts:    ctx.getRmConfigOpts(),
	}, "config", "rm")
	root.AddChainAnyOrder(&Command{
		Desc:    "Create new config.",
		Handler: ctx.NewConfigCommand,
		Opts:    newConfigOpts,
	}, "config", "new")

	ctx.CommandRoot = root

	return ctx, nil
}
