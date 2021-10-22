package cmd

import (
	"fmt"
	"gitea-cli/config"
	"gitea-cli/gitea"
	"os"
	"os/exec"
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

func newConfigOpts(c *config.Config) []CmdOpt {

	var (
		username, url, tn, ver, repo, owner, base string
	)

	if c == nil {
		c = &config.Config{}
	}

	if c.Username != "" {
		username = fmt.Sprintf(" [empty for '%s']", c.Username)
	}
	if c.RepoUrl != "" {
		url = fmt.Sprintf(" [empty for '%s']", c.RepoUrl)
	}
	if c.TokenName != "" {
		tn = fmt.Sprintf(" [empty for '%s']", c.TokenName)
	} else {
		tn = " [empty for rand]"
	}
	if c.ApiVer != "" {
		ver = fmt.Sprintf(" [empty for '%s']", c.ApiVer)
	} else {
		ver = " empty for 'v1'"
	}
	if c.DefaultRepoOwner != "" {
		owner = fmt.Sprintf(" [empty for '%s']", c.DefaultRepoOwner)
	}
	if c.DefaultRepoName != "" {
		repo = fmt.Sprintf(" [empty for '%s']", c.DefaultRepoName)
	}
	if c.DefaultBaseForMr != "" {
		base = fmt.Sprintf(" [empty for '%s']", c.DefaultBaseForMr)
	}

	return []CmdOpt{
		{ // 0
			Spec: CmdOptSpec{
				ArgFlags: []string{"u", "username"},
				Label:    "Your gitea username" + username,
				Optional: username != "",
				DefaultStrFunc: func() (string, error) {
					return c.Username, nil
				},
			},
		}, { // 1
			Spec: CmdOptSpec{
				Label:    "Your gitea password",
				NoEcho:   true,
				NoPrompt: c.TokenSha1 != "",
			},
		}, { // 2
			Spec: CmdOptSpec{
				ArgFlags: []string{"g", "gitea"},
				Label:    "Gitea server url" + url,
				Optional: url != "",
				DefaultStrFunc: func() (string, error) {
					return c.RepoUrl, nil
				},
			},
		}, { // 3
			Spec: CmdOptSpec{
				ArgFlags: []string{"n", "name"},
				Label:    "Token name" + tn,
				DefaultStrFunc: func() (string, error) {
					if c.TokenName != "" {
						return c.TokenName, nil
					}
					return randStr(8), nil
				},
				Optional: true,
			},
		}, { // 4
			Spec: CmdOptSpec{
				ArgFlags: []string{"V", "apiver"},
				Label:    "Gitea api version" + ver,
				NoPrompt: true,
				Optional: true,
				DefaultStrFunc: func() (string, error) {
					if c.ApiVer != "" {
						return c.ApiVer, nil
					}
					return "v1", nil
				},
			},
		}, { // 5
			Spec: CmdOptSpec{
				ArgFlags: []string{"o", "owner"},
				Label:    "Default repo owner (used in PRs)" + owner,
				Optional: true,
				DefaultStrFunc: func() (string, error) {
					return c.DefaultRepoOwner, nil
				},
			},
		}, { // 6
			Spec: CmdOptSpec{
				ArgFlags: []string{"r", "repo"},
				Label:    "Default repo name (used in PRs)" + repo,
				Optional: true,
				DefaultStrFunc: func() (string, error) {
					return c.DefaultRepoName, nil
				},
			},
		}, { // 7
			Spec: CmdOptSpec{
				ArgFlags: []string{"b", "base"},
				Label:    "Default base branch (target branch in PRs)" + base,
				Optional: true,
				DefaultStrFunc: func() (string, error) {
					return c.DefaultBaseForMr, nil
				},
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
}

func (ctx *CmdCtx) NewConfigCommand() error {

	newConfigOpts := newConfigOpts(ctx.Config)
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

	var token *gitea.Token
	// if token already exists - dont recreate it
	if ctx.Config == nil || ctx.Config.TokenSha1 == "" {
		var err error
		token, err = tr.GetToken()
		if err != nil {
			return err
		}
		fmt.Printf("Created gitea token with name: %s\n", token.Name)
	} else {
		token = &gitea.Token{
			Sha1: ctx.Config.TokenSha1,
			Name: ctx.Config.TokenName,
		}
		fmt.Printf("Using gitea token with name: %s\n", token.Name)
	}

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
			NoPrompt: optional,
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
	head := getBranch()
	ret = append(ret, addOptWithDefaultVal("head branch", "source branch in PR", []string{"h", "head"}, head))

	// 3
	def = ""
	if config != nil && config.DefaultBaseForMr != "" {
		def = config.DefaultBaseForMr
	}
	ret = append(ret, addOptWithDefaultVal("base branch", "target branch in PR", []string{"b", "base"}, def))

	// 4
	ret = append(ret, addOptWithDefaultVal("pr title   ", "", []string{"t", "title"}, head))

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

	if title == "" {
		title = head
	}

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
			Title: title,
		},
	}
	pr, err := repoCtx.CreatePR(&req)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", pr.Url)

	return nil
}

func findPrOpts() []CmdOpt {
	return []CmdOpt{
		{
			Spec: CmdOptSpec{
				ArgFlags: []string{"t", "title"},
				Label:    "PR title (current branch name if empty)",
				DefaultStrFunc: func() (string, error) {
					return getBranch(), nil
				},
				NoPrompt: true,
			},
		},
	}
}

func findPr(repoCtx *gitea.RepoCtx, title string) (gitea.PullRequest, error) {
	req := gitea.ListPRRequest{
		State: "open",
	}
	prs, err := repoCtx.ListPR(&req)
	if err != nil {
		return gitea.PullRequest{}, err
	}

	for i := range prs {
		if prs[i].Title == title {
			return prs[i], nil
		}
	}

	return gitea.PullRequest{}, fmt.Errorf("pr not found")
}

func mergePrOpts(config *config.Config) []CmdOpt {
	opts := repoInfoOpts(config)
	opts = append(opts, findPrOpts()...)
	opts = append(opts, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"rm", "del"},
			Label:    "Remove branch",
			Optional: true,
			NoPrompt: true,
			IsBool:   true,
		},
	})
	opts = append(opts, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"f", "force"},
			Label:    "Force merge",
			Optional: true,
			NoPrompt: true,
			IsBool:   true,
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
	title := opts[2].Val.Str
	rm := opts[3].Val.Bool
	force := opts[4].Val.Bool

	var err error

	repoCtx := gitea.RepoCtx{
		Token:  ctx.Config.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.ToRepoApiUrl(),
	}

	fmt.Printf("merging pr with title: '%s'\n", title)

	pr, err := findPr(&repoCtx, title)
	if err != nil {
		return err
	}
	index := pr.Number

	mergeReq := gitea.MergePRRequest{
		Opt: gitea.MergePullRequestOption{
			Do:         "squash",
			ForceMerge: force,
		},
		Index: index,
	}
	if err := repoCtx.MergePR(&mergeReq); err != nil {
		return err
	}

	if !rm {
		var tmp string
	O:
		for {
			fmt.Printf("remove branch %s? [y/n]: ", pr.Head.Ref)
			fmt.Scanln(&tmp)
			switch tmp {
			case "y":
				rm = true
				break O
			case "n":
				break O
			}
		}
	}

	if rm {
		return repoCtx.DeleteBranch(&gitea.DeleteBranchRequest{
			Branch: pr.Head.Ref,
		})
	}

	return nil
}

//

func updatePrOpts(config *config.Config) []CmdOpt {
	opts := repoInfoOpts(config)
	opts = append(opts, findPrOpts()...)

	opts = append(opts, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"c", "close"},
			Label:    "close pull request",
			IsBool:   true,
			Optional: true,
			NoPrompt: true,
		},
	})

	opts = append(opts, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"rename"},
			Label:    "Change title",
			IsBool:   false,
			Optional: true,
			NoPrompt: true,
		},
	})

	return opts
}

func (ctx *CmdCtx) ClosePrCommand() error {

	if err := ctx.hasConfig(); err != nil {
		return err
	}

	opts := updatePrOpts(ctx.Config)
	if err := GetOpts(os.Args[1:], opts); err != nil {
		return err
	}

	owner := opts[0].Val.Str
	repo := opts[1].Val.Str
	title := opts[2].Val.Str

	close := opts[3].Val.Bool
	rename := opts[4].Val.Str

	var err error

	repoCtx := gitea.RepoCtx{
		Token:  ctx.Config.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.ToRepoApiUrl(),
	}

	fmt.Printf("updating pr with title: '%s'\n", title)

	pr, err := findPr(&repoCtx, title)
	if err != nil {
		return err
	}
	index := pr.Number

	req := gitea.UpdatePrRequest{
		Index: index,
		Opt: gitea.EditPullRequestOption{
			Title: rename,
		},
	}

	if close {
		req.Opt.State = gitea.Closed
	}

	if err := repoCtx.UpdatePR(&req); err != nil {
		return err
	}

	return nil
}
