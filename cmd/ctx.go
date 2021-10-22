package cmd

import (
	"fmt"
	"gitea-cli/common"
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
	Config      *common.Config
}

func (ctx *CmdCtx) ValidateConfig(withCred bool) error {
	if ctx.Config == nil {
		return fmt.Errorf("config is nil")
	}
	return ctx.Config.Validate(withCred)
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
	const indent = "  "

	fmt.Print("Manage gitea repository\n\n")

	fmt.Printf("Usage: \n%s%s [options] [command]\n%s%s [command] [options]\n\n",
		indent, os.Args[0], indent, os.Args[0])

	fmt.Print("Commands:\n\n")

	for c := range gc {
		//fmt.Printf("%s%s\n%sUsage:\n", indent, c, strings.Repeat(indent, 2))
		for i := range gc[c].Paths {
			fmt.Print(gc[c].Paths[i])
			if i == len(gc[c].Paths)-1 {
				fmt.Printf("  \t%s", c)
			}
			fmt.Println()
		}

		if gc[c].Command.Opts != nil {

			for i := range gc[c].Command.Opts {
				o := gc[c].Command.Opts[i]
				if len(o.Spec.ArgFlags) == 0 {
					continue
				}
				if i == 0 {
					fmt.Printf("%sArguments:\n", indent)
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
		fmt.Print("\n")
	}
}

// create context using saved config
func NewCtx() (*CmdCtx, error) {
	ctx := new(CmdCtx)

	fc, err := ioutil.ReadFile("gitea.yml")
	if err != nil {
		return nil, err
	} else {
		ctx.Config = new(common.Config)
		if err := yaml.Unmarshal(fc, ctx.Config); err != nil {
			return nil, err
		}
	}

	root := Branch{}

	root.AddChainAnyOrder(&Command{
		Desc:    "",
		Handler: ctx.HelpCommand,
	}, "help")
	root.AddChainStrictOrder(&Command{
		Desc:    "Create new gitea credentials.",
		Handler: ctx.NewGiteaCredCommand,
		Opts:    newGiteaCredOpts(),
	}, "new", "g", "cred")
	root.AddChainStrictOrder(&Command{
		Desc:    "Create new rocketchat credentials.",
		Handler: ctx.NewRocketCredCommand,
		Opts:    ctx.newRocketCredOpts(true),
	}, "new", "r", "cred")
	root.AddChainStrictOrder(&Command{
		Desc:    "Create new credentials.",
		Handler: ctx.NewCredCommand,
		Opts:    ctx.newCredOpts(),
	}, "new", "cred")
	root.AddChainStrictOrder(&Command{
		Desc:    "Remove credentials.",
		Handler: ctx.RmCredCommand,
		Opts:    ctx.getRmCredOpts(),
	}, "rm", "cred")
	root.AddChainStrictOrder(&Command{
		Desc:    "Create new pull request.",
		Handler: ctx.NewPrCommand,
		Opts:    newPrOpts(ctx.Config),
	}, "new", "pr")
	root.AddChainStrictOrder(&Command{
		Desc:    "list open pull requests.",
		Handler: ctx.ListPrCommand,
		Opts:    listPrOpts(ctx.Config),
	}, "list", "pr")
	root.AddChainStrictOrder(&Command{
		Desc:    "Merge existing pull request",
		Handler: ctx.MergePrCommand,
		Opts:    mergePrOpts(ctx.Config),
	}, "merge", "pr")
	root.AddChainStrictOrder(&Command{
		Desc:    "Close existing pull request",
		Handler: ctx.ClosePrCommand,
		Opts:    updatePrOpts(ctx.Config),
	}, "update", "pr")

	ctx.CommandRoot = root

	return ctx, nil
}
