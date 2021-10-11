package cmd

import (
	"fmt"
	"gitea-cli/config"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

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

	ctx.Commands = map[string]*Command{
		"config": {
			Name:    "config",
			Desc:    "manage local config",
			Handler: ConfigCmd,
		},
		"help": {
			Name:    "help",
			Desc:    "list available commands",
			Handler: HelpCmd,
		},
		"pr": {
			Name:    "pr",
			Desc:    "manage pull requests",
			Handler: PRCmd,
		},
	}
	return ctx, nil
}

type Command struct {
	Name    string
	Desc    string
	Handler func(ctx *CmdCtx, argv []string) error
}

type CmdCtx struct {
	Commands map[string]*Command
	Config   *config.Config
}
