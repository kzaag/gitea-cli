package cmd

import (
	"fmt"
	"gitea-cli/config"
	"gitea-cli/gitea"
	"os"

	"gopkg.in/yaml.v2"
)

func (ctx *CmdCtx) HelpCommand() error {
	ctx.PrintCommands()
	return nil
}

func (ctx *CmdCtx) getRmConfigOpts() []CmdOpt {
	s := ""
	if ctx.Config != nil {
		s = fmt.Sprintf("Password for %s @ %s", ctx.Config.Username, ctx.Config.RepoUrl)
	} else {
		s = "Password for your gitea user"
	}
	return []CmdOpt{
		{
			Spec: CmdOptSpec{
				Label:       s,
				NotRequired: false,
				NoEcho:      true,
				IsBool:      false,
			},
		},
	}
}

func (ctx *CmdCtx) RmConfigCommand() error {
	if ctx.Config == nil {
		return fmt.Errorf("delete config: config not found")
	}

	opts := ctx.getRmConfigOpts()
	GetOpts(os.Args[1:], opts)

	c := gitea.TokenRequest{
		Username: ctx.Config.Username,
		Password: opts[0].Val.Str,
		TokenRequestBody: gitea.TokenRequestBody{
			TokenName: ctx.Config.TokenName,
		},
		RepoInfo: ctx.Config.RepoInfo,
	}
	if err := c.DeleteToken(ctx.Config.TokenName); err != nil {
		return err
	}
	return os.Remove("gitea.yml")
}

var newConfigOpts = []CmdOpt{
	{ // 0
		Spec: CmdOptSpec{
			ArgFlags: []string{"u", "username"},
			Label:    "Your gitea username",
		},
	}, { // 1
		Spec: CmdOptSpec{
			Label:  "Your gitea password",
			NoEcho: true,
		},
	}, { // 2
		Spec: CmdOptSpec{
			ArgFlags: []string{"g", "gitea"},
			Label:    "Gitea server url (https://gitea.com/relative/path)",
		},
	}, { // 3
		Spec: CmdOptSpec{
			ArgFlags: []string{"n", "name"},
			Label:    "Token name (empty for rand)",
			DefaultStrFunc: func() string {
				return randStr(8)
			},
			NotRequired: true,
		},
	}, { // 4
		Spec: CmdOptSpec{
			ArgFlags:    []string{"V", "apiver"},
			Label:       "Gitea api version, default v1",
			NoPrompt:    true,
			NotRequired: true,
			DefaultStrFunc: func() string {
				return "v1"
			},
		},
	}, { // 5
		Spec: CmdOptSpec{
			ArgFlags:    []string{"o", "owner"},
			Label:       "Default repo owner (used in PRs) [empty if none]",
			NotRequired: true,
		},
	}, { // 6
		Spec: CmdOptSpec{
			ArgFlags:    []string{"r", "repo"},
			Label:       "Default repo name (used in PRs) [empty if none]",
			NotRequired: true,
		},
	}, { // 7
		Spec: CmdOptSpec{
			ArgFlags:    []string{"b", "base"},
			Label:       "Default base branch (target branch in PRs) [empty if none]",
			NotRequired: true,
		},
	}, { // 8
		Spec: CmdOptSpec{
			ArgFlags: []string{"p", "path"},
			Label:    "Config path [empty for ./gitea.yml]",
			NoPrompt: true,
			DefaultStrFunc: func() string {
				return "./gitea.yml"
			},
			NotRequired: true,
		},
	},
}

func (ctx *CmdCtx) NewConfigCommand() error {

	GetOpts(os.Args[1:], newConfigOpts)

	tr := gitea.TokenRequest{
		Username: newConfigOpts[0].Val.Str,
		Password: newConfigOpts[1].Val.Str,
		RepoInfo: config.RepoInfo{
			ApiVer:  newConfigOpts[4].Val.Str,
			RepoUrl: newConfigOpts[2].Val.Str,
		},
		TokenRequestBody: gitea.TokenRequestBody{
			TokenName: newConfigOpts[3].Val.Str,
		},
	}

	if err := tr.Validate(); err != nil {
		return err
	}

	token, err := tr.GetToken()
	if err != nil {
		return err
	}

	fmt.Printf("Created gitea token with name: %s\n", token.Name)

	conf := config.Config{
		TokenSha1:        token.Sha1,
		TokenName:        token.Name,
		Username:         tr.Username,
		DefaultRepoName:  newConfigOpts[6].Val.Str,
		DefaultRepoOwner: newConfigOpts[5].Val.Str,
		DefaultBaseForMr: newConfigOpts[7].Val.Str,
		RepoInfo:         tr.RepoInfo,
	}

	fp, err := os.OpenFile(newConfigOpts[8].Val.Str, os.O_CREATE|os.O_WRONLY, 0600)
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

func (ctx *CmdCtx) NewPrCommand() {

}
