package cmd

import (
	"fmt"
	"gitea-cli/config"
	"gitea-cli/gitea"
	"os"

	"gopkg.in/yaml.v2"
)

func DelLocalConfig(ctx *CmdCtx) error {
	if ctx.Config == nil {
		return fmt.Errorf("delete config: config not found")
	}
	var c gitea.TokenRequest
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

func NewLocalConfig(ctx *CmdCtx) error {
	var c gitea.TokenRequest
	if err := c.FillFromConsole(); err != nil {
		return err
	}
	t, err := c.GetToken()
	if err != nil {
		return err
	}
	fmt.Printf("Created gitea token with name: %s\n", t.Name)

	var conf config.Config
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

func printConfigHelp() {
	fmt.Println(GetHelpStr([]NameWithDesc{
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
}

func ConfigCmd(ctx *CmdCtx, argv []string) error {
	cmd := ""
	if len(argv) == 0 {
		cmd = "new"
	} else {
		cmd = argv[0]
	}

	switch cmd {
	case "help":
		printConfigHelp()
	case "new":
		return NewLocalConfig(ctx)
	case "del":
		return DelLocalConfig(ctx)
	default:
		printConfigHelp()
		return fmt.Errorf("ConfigCmd: '%s' not found", cmd)
	}

	return nil
}
