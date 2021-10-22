package cmd

import (
	"gitea-cli/gitea"
	"gitea-cli/rocketchat"
	"os"

	"gopkg.in/yaml.v2"
)

func (ctx *CmdCtx) getRmCredOpts() []CmdOpt {
	return []CmdOpt{
		{ // 0
			Spec: CmdOptSpec{
				ArgFlags: []string{"u", "user"},
				Label:    "Gitea username",
			},
		}, { // 1
			Spec: CmdOptSpec{
				Label:  "Gitea password",
				NoEcho: true,
			},
		},
	}
}

func (ctx *CmdCtx) PersistConfig() error {
	fp, err := os.OpenFile("./gitea.yml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()

	b, err := yaml.Marshal(&ctx.Config)
	if err != nil {
		return err
	}

	if _, err := fp.Write(b); err != nil {
		return err
	}

	return nil
}

func (ctx *CmdCtx) RmCredCommand() error {

	if err := ctx.ValidateConfig(true); err != nil {
		return err
	}

	rmCredOpts := ctx.getRmCredOpts()

	GetOpts(os.Args[1:], rmCredOpts)

	guser := rmCredOpts[0].Val.Str
	gpass := rmCredOpts[1].Val.Str

	giteaReq := gitea.TokenRequest{
		Username: guser,
		Password: gpass,
		TokenRequestBody: gitea.TokenRequestBody{
			TokenName: ctx.Config.Gitea.TokenName,
		},
		RemoteInfo: ctx.Config.Gitea.RemoteInfo,
	}

	if err := giteaReq.DeleteToken(); err != nil {
		return err
	}

	ctx.Config.Gitea.TokenName = ""
	ctx.Config.Gitea.TokenSha1 = ""

	return ctx.PersistConfig()
}

func (ctx *CmdCtx) newRocketCredOpts(withRocketchat bool) []CmdOpt {
	return []CmdOpt{
		{
			Spec: CmdOptSpec{
				ArgFlags: []string{"rocketuser"},
				Label:    "Rocketchat username",
				NoPrompt: !withRocketchat,
				Optional: !withRocketchat,
			},
		}, {
			Spec: CmdOptSpec{
				Label:    "Rocketchat password",
				NoEcho:   true,
				NoPrompt: !withRocketchat,
				Optional: !withRocketchat,
			},
		},
	}
}

func newGiteaCredOpts() []CmdOpt {
	return []CmdOpt{
		{ // 0
			Spec: CmdOptSpec{
				ArgFlags: []string{"u", "user"},
				Label:    "Gitea username",
			},
		}, { // 1
			Spec: CmdOptSpec{
				Label:  "Gitea password",
				NoEcho: true,
			},
		}, { // 2
			Spec: CmdOptSpec{
				ArgFlags: []string{"t", "token"},
				Label:    "Token name (default: random)",
				DefaultStrFunc: func() (string, error) {
					return randStr(8), nil
				},
				Optional: true,
			},
		},
	}
}

func (ctx *CmdCtx) newCredOpts() []CmdOpt {

	withRocketchat := false
	if ctx.Config != nil && ctx.Config.Rocketchat.Enabled {
		withRocketchat = true
	}

	ret := newGiteaCredOpts()

	ret = append(ret, ctx.newRocketCredOpts(withRocketchat)...)

	return ret
}

func (ctx *CmdCtx) setNewGiteaCred(opts []CmdOpt) error {

	guser := opts[0].Val.Str
	gpass := opts[1].Val.Str
	gtoken := opts[2].Val.Str

	giteaReq := gitea.TokenRequest{
		Username: guser,
		Password: gpass,
		TokenRequestBody: gitea.TokenRequestBody{
			TokenName: gtoken,
		},
		RemoteInfo: ctx.Config.Gitea.RemoteInfo,
	}

	giteaToken, err := giteaReq.GetToken()
	if err != nil {
		return err
	}

	ctx.Config.Gitea.TokenName = gtoken
	ctx.Config.Gitea.TokenSha1 = giteaToken.Sha1

	return nil
}

func (ctx *CmdCtx) setNewRocketchatCred(opts []CmdOpt) error {
	ruser := opts[0].Val.Str
	rpass := opts[1].Val.Str
	rocketReq := rocketchat.LoginRequest{
		User:     ruser,
		Password: rpass,
	}
	rocketRes, err := rocketchat.Login(&ctx.Config.Rocketchat.RemoteInfo, &rocketReq)
	if err != nil {
		return err
	}
	ctx.Config.Rocketchat.Token = rocketRes.Data.AuthToken
	ctx.Config.Rocketchat.UserID = rocketRes.Data.UserID

	return nil
}

func (ctx *CmdCtx) NewRocketCredCommand() error {

	if err := ctx.ValidateConfig(false); err != nil {
		return err
	}

	newCredOpts := ctx.newRocketCredOpts(true)
	GetOpts(os.Args[1:], newCredOpts)

	if err := ctx.setNewRocketchatCred(newCredOpts); err != nil {
		return err
	}

	return ctx.PersistConfig()
}

func (ctx *CmdCtx) NewGiteaCredCommand() error {

	if err := ctx.ValidateConfig(false); err != nil {
		return err
	}

	newCredOpts := newGiteaCredOpts()
	GetOpts(os.Args[1:], newCredOpts)

	if err := ctx.setNewGiteaCred(newCredOpts); err != nil {
		return err
	}

	return ctx.PersistConfig()
}

func (ctx *CmdCtx) NewCredCommand() error {

	if err := ctx.ValidateConfig(false); err != nil {
		return err
	}

	newCredOpts := ctx.newCredOpts()
	GetOpts(os.Args[1:], newCredOpts)

	if err := ctx.setNewGiteaCred(newCredOpts); err != nil {
		return err
	}

	if ctx.Config.Rocketchat.Enabled {
		if err := ctx.setNewRocketchatCred(newCredOpts[3:]); err != nil {
			return err
		}
	}

	return ctx.PersistConfig()
}
