package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

const RepoApiUrl = "https://git.mulwi.cloud/gitrepos/api/v1"

type Config struct {
	TokenSha1 string `yaml:"token_sha1"`
	TokenName string `yaml:"token_name"`
	Username  string `yaml:"username"`
	// Default repository, can be empty
	DefaultRepoName  string `yaml:"default_repo_name"`
	DefaultRepoOwner string `yaml:"default_repo_owner"`
}

func (c *Config) validationErr(msg string) error {
	return fmt.Errorf("Validate Config: %s", msg)
}

func (c *Config) Validate() error {
	if c.TokenSha1 == "" {
		return c.validationErr("invalid token_sha1")
	}
	if c.TokenName == "" {
		return c.validationErr("invalid token_name")
	}
	if c.Username == "" {
		return c.validationErr("invalid username")
	}
	return nil
}

type Command struct {
	Name    string
	Desc    string
	Handler func(ctx *AppCtx, argv []string) error
}

type AppCtx struct {
	Commands map[string]*Command
	Config   *Config
}

// create context using saved config
func NewCtx() (*AppCtx, error) {
	ctx := new(AppCtx)

	fc, err := ioutil.ReadFile("gitea.yml")
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		cnf := new(Config)
		if err := yaml.Unmarshal(fc, cnf); err != nil {
			return nil, err
		}
		if err := cnf.Validate(); err != nil {
			return nil, err
		}
		fmt.Printf("Using config for user %s\n", cnf.Username)
		ctx.Config = cnf
	}

	ctx.Commands = map[string]*Command{
		"config": {
			Name:    "config",
			Desc:    "manage local config",
			Handler: ConfigHandler,
		},
		"help": {
			Name:    "help",
			Desc:    "list available commands",
			Handler: HelpHandler,
		},
	}
	return ctx, nil
}
