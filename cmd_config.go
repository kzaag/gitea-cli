package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func DelLocalConfig(ctx *AppCtx) error {
	if ctx.Config == nil {
		return fmt.Errorf("delete config: config not found")
	}
	var c TokenRequest
	c.TokenName = ctx.Config.TokenName
	c.Username = ctx.Config.Username
	if err := c.FillFromConsole(); err != nil {
		return err
	}
	if err := c.DeleteToken(ctx.Config.TokenName); err != nil {
		return err
	}
	return os.Remove("gitea.yml")
}

func NewLocalConfig(ctx *AppCtx) error {
	var c TokenRequest
	if err := c.FillFromConsole(); err != nil {
		return err
	}
	t, err := c.GetToken()
	if err != nil {
		return err
	}
	fmt.Printf("Created gitea token with name: %s\n", t.Name)

	var conf Config
	conf.TokenName = t.Name
	conf.TokenSha1 = t.Sha1
	conf.Username = c.Username
	conf.RepoInfo = c.RepoInfo

	fmt.Print("Default repo owner: [empty if none]: ")
	fmt.Scanln(&conf.DefaultRepoOwner)
	fmt.Print("Default repo name: [empty if none]: ")
	fmt.Scanln(&conf.DefaultRepoName)
	fmt.Print("Default base branch: [empty if none]: ")
	fmt.Scanln(&conf.DefaultBaseForMr)

	fp, err := os.OpenFile("gitea.yml", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()

	b, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}

	if _, err := fp.Write(b); err != nil {
		return err
	}

	return nil
}

func ConfigHandler(ctx *AppCtx, argv []string) error {
	cmd := ""
	if len(argv) < 2 {
		cmd = "new"
	}

	if cmd == "" {
		cmd = argv[1]
	}

	switch cmd {
	case "help":
		fmt.Println(GetHelpStr([]NameDesc{
			{
				Name: "help",
				Desc: "prints info",
			},
			{
				Name: "new",
				Desc: "[default]. create config and save it for later use",
			},
			{
				Name: "del",
				Desc: "delete and remove config",
			},
		}))
	case "new":
		return NewLocalConfig(ctx)
	case "del":
		return DelLocalConfig(ctx)
	default:
		return fmt.Errorf("Token: '%s' not found, Use 'token help' to list available commands", argv[1])
	}

	return nil
}
