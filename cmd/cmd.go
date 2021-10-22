package cmd

import (
	"fmt"
	"gitea-cli/config"
	"gitea-cli/gitea"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
				Label:    s,
				Optional: false,
				NoEcho:   true,
				IsBool:   false,
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
			DefaultStrFunc: func() (string, error) {
				return randStr(8), nil
			},
			Optional: true,
		},
	}, { // 4
		Spec: CmdOptSpec{
			ArgFlags: []string{"V", "apiver"},
			Label:    "Gitea api version, default v1",
			NoPrompt: true,
			Optional: true,
			DefaultStrFunc: func() (string, error) {
				return "v1", nil
			},
		},
	}, { // 5
		Spec: CmdOptSpec{
			ArgFlags: []string{"o", "owner"},
			Label:    "Default repo owner (used in PRs) [empty if none]",
			Optional: true,
		},
	}, { // 6
		Spec: CmdOptSpec{
			ArgFlags: []string{"r", "repo"},
			Label:    "Default repo name (used in PRs) [empty if none]",
			Optional: true,
		},
	}, { // 7
		Spec: CmdOptSpec{
			ArgFlags: []string{"b", "base"},
			Label:    "Default base branch (target branch in PRs) [empty if none]",
			Optional: true,
		},
	}, { // 8
		Spec: CmdOptSpec{
			ArgFlags: []string{"p", "path"},
			Label:    "Config path [empty for ./gitea.yml]",
			NoPrompt: true,
			DefaultStrFunc: func() (string, error) {
				return "./gitea.yml", nil
			},
			Optional: true,
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

func addOptWithDefaultVal(label, helptext string, flags []string, defaultValue string) CmdOpt {
	var (
		optional bool
	)

	if defaultValue != "" {
		optional = true
		label = fmt.Sprintf("%s [empty for: '%s']", label, defaultValue)
	} else {
		optional = false
		if helptext != "" {
			label = fmt.Sprintf("%s (%s)", label, helptext)
		}
		defaultValue = ""
	}

	return CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: flags,
			Label:    label,
			Optional: optional,
			DefaultStrFunc: func() (string, error) {
				return defaultValue, nil
			},
		},
	}
}

func getBranch() string {
	o, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(o))
}

func repoInfoOpts(config *config.Config) []CmdOpt {
	ret := make([]CmdOpt, 0, 4)

	var (
		def string
	)

	// 0
	def = ""
	if config != nil && config.DefaultRepoOwner != "" {
		def = config.DefaultRepoOwner
	}
	ret = append(ret, addOptWithDefaultVal("repo owner ", "", []string{"o", "owner"}, def))

	// 1
	def = ""
	if config != nil && config.DefaultRepoName != "" {
		def = config.DefaultRepoName
	}
	ret = append(ret, addOptWithDefaultVal("repo name  ", "", []string{"r", "repo"}, def))

	return ret
}

func (ctx *CmdCtx) hasConfig() error {
	if ctx.Config == nil {
		return fmt.Errorf("config not found. Execute '%s new config' to create config", os.Args[0])
	}
	return nil
}

func listPrOpts(config *config.Config) []CmdOpt {
	return repoInfoOpts(config)
}

func (ctx *CmdCtx) ListPrCommand() error {
	if err := ctx.hasConfig(); err != nil {
		return err
	}

	opts := listPrOpts(ctx.Config)
	if err := GetOpts(os.Args[1:], opts); err != nil {
		return err
	}
	owner := opts[0].Val.Str
	repo := opts[1].Val.Str

	repoCtx := gitea.RepoCtx{
		Token:  ctx.Config.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.ToRepoApiUrl(),
	}

	req := gitea.ListPRRequest{}
	req.State = "open"
	prs, err := repoCtx.ListPR(&req)
	if err != nil {
		return err
	}

	for i := range prs {
		fmt.Printf("PR: %s->%s index=%d, title=%s, user=%s, url=%s\n",
			prs[i].Head.Ref, prs[i].Base.Ref,
			prs[i].Number, prs[i].Title, prs[i].User.Login,
			prs[i].Url)
	}

	return nil
}

func newPrOpts(config *config.Config) []CmdOpt {

	ret := make([]CmdOpt, 0, 7)

	var (
		def string
	)

	// 0, 1
	ret = append(ret, repoInfoOpts(config)...)

	// 2
	def = getBranch()
	ret = append(ret, addOptWithDefaultVal("head branch", "source branch in PR", []string{"h", "head"}, def))

	// 3
	def = ""
	if config != nil && config.DefaultBaseForMr != "" {
		def = config.DefaultBaseForMr
	}
	ret = append(ret, addOptWithDefaultVal("base branch", "target branch in PR", []string{"b", "base"}, def))

	// 4
	ret = append(ret, addOptWithDefaultVal("pr title   ", "", []string{"t", "title"}, def))

	// 5
	ret = append(ret, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"w", "wip"},
			Label:    "work in progress [default: false]",
			NoPrompt: true,
			IsBool:   true,
		},
	})

	// 6
	ret = append(ret, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"d", "dry"},
			Label:    "dry run [default: false]",
			NoPrompt: true,
			IsBool:   true,
		},
	})

	return ret
}

func (ctx *CmdCtx) NewPrCommand() error {
	if err := ctx.hasConfig(); err != nil {
		return err
	}

	opts := newPrOpts(ctx.Config)
	if err := GetOpts(os.Args[1:], opts); err != nil {
		return err
	}

	owner := opts[0].Val.Str
	repo := opts[1].Val.Str
	head := opts[2].Val.Str
	base := opts[3].Val.Str
	title := opts[4].Val.Str
	wip := opts[5].Val.Bool
	dry := opts[6].Val.Bool

	if wip {
		title = "WIP: " + title
	}

	fmt.Printf("Creating pr for %s/%s %s->%s with title: '%s'\n", owner, repo, head, base, title)

	if dry {
		return nil
	}

	repoCtx := gitea.RepoCtx{
		Token:  ctx.Config.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.ToRepoApiUrl(),
	}
	if err := repoCtx.Validate(); err != nil {
		return err
	}

	var req = gitea.CreatePRRequest{
		Opt: gitea.CreatePullRequestOption{
			Base:  base,
			Head:  head,
			Title: head,
		},
	}
	pr, err := repoCtx.CreatePR(&req)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", pr.Url)

	return nil
}

func mergePrOpts(config *config.Config) []CmdOpt {
	opts := repoInfoOpts(config)

	// 2
	opts = append(opts, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"i", "index"},
			Label:    "PR index",
			Optional: false,
		},
	})

	return opts
}

func (ctx *CmdCtx) MergePrCommand() error {

	if err := ctx.hasConfig(); err != nil {
		return err
	}

	opts := mergePrOpts(ctx.Config)
	if err := GetOpts(os.Args[1:], opts); err != nil {
		return err
	}

	owner := opts[0].Val.Str
	repo := opts[1].Val.Str
	index := opts[2].Val.Str

	indexInt, err := strconv.Atoi(index)
	if err != nil {
		return err
	}

	repoCtx := gitea.RepoCtx{
		Token:  ctx.Config.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.ToRepoApiUrl(),
	}
	mergeReq := gitea.MergePRRequest{
		Opt: gitea.MergePullRequestOption{
			Do:         "squash",
			ForceMerge: true,
		},
		Index: indexInt,
	}
	if err := repoCtx.MergePR(&mergeReq); err != nil {
		return err
	}

	return nil
}
