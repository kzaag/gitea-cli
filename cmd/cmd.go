package cmd

import (
	"fmt"
	"gitea-cli/common"
	"gitea-cli/gitea"
	"gitea-cli/rocketchat"
	"os"
	"os/exec"
	"strings"
)

func (ctx *CmdCtx) HelpCommand() error {
	ctx.PrintCommands()
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

func repoInfoOpts(config *common.Config) []CmdOpt {
	ret := make([]CmdOpt, 0, 4)

	var (
		def string
	)

	// 0
	def = ""
	if config != nil && config.Gitea.DefaultRepoOwner != "" {
		def = config.Gitea.DefaultRepoOwner
	}
	ret = append(ret, addOptWithDefaultVal("repo owner ", "", []string{"o", "owner"}, def))

	// 1
	def = ""
	if config != nil && config.Gitea.DefaultRepoName != "" {
		def = config.Gitea.DefaultRepoName
	}
	ret = append(ret, addOptWithDefaultVal("repo name  ", "", []string{"r", "repo"}, def))

	return ret
}

func listPrOpts(config *common.Config) []CmdOpt {
	return repoInfoOpts(config)
}

func (ctx *CmdCtx) ListPrCommand() error {
	if err := ctx.ValidateConfig(true); err != nil {
		return err
	}

	opts := listPrOpts(ctx.Config)
	if err := GetOpts(os.Args[1:], opts); err != nil {
		return err
	}
	owner := opts[0].Val.Str
	repo := opts[1].Val.Str

	repoCtx := gitea.RepoCtx{
		Token:  ctx.Config.Gitea.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.Gitea.ToApiUrl(),
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

func newPrOpts(c *common.Config) []CmdOpt {

	ret := make([]CmdOpt, 0, 7)

	var (
		def string
	)

	// 0, 1
	ret = append(ret, repoInfoOpts(c)...)

	// 2
	head := getBranch()
	ret = append(ret, addOptWithDefaultVal("head branch", "source branch in PR", []string{"h", "head"}, head))

	// 3
	def = ""
	if c != nil && c.Gitea.DefaultBaseForPR != "" {
		def = c.Gitea.DefaultBaseForPR
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

	// 7
	optional := !c.Rocketchat.Enabled || c.Rocketchat.DefaultNotifyChannel != ""
	ret = append(ret, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"notify"},
			Label:    "Notification channel for rocketchat",
			NoPrompt: optional,
			Optional: optional,
			DefaultStrFunc: func() (string, error) {
				return c.Rocketchat.DefaultNotifyChannel, nil
			},
		},
	})

	// 8
	ret = append(ret, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"n", "nohdr"},
			Label:    "dont use default_header field from config [default: false]",
			NoPrompt: true,
			IsBool:   true,
		},
	})

	// 9
	ret = append(ret, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"f", "footer"},
			Label:    "text below message [default: empty]",
			NoPrompt: true,
			DefaultStrFunc: func() (string, error) {
				return "", nil
			},
		},
	})

	return ret
}

func (ctx *CmdCtx) NewPrCommand() error {
	if err := ctx.ValidateConfig(true); err != nil {
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
		Token:  ctx.Config.Gitea.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.Gitea.ToApiUrl(),
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

	if ctx.Config.Rocketchat.Enabled {
		rctx := rocketchat.Ctx{
			ApiUrl: ctx.Config.Rocketchat.ToApiUrl(),
			UserID: ctx.Config.Rocketchat.UserID,
			Token:  ctx.Config.Rocketchat.Token,
		}

		targetChan := opts[7].Val.Str

		noHdr := opts[8].Val.Bool
		msg := ""

		if ctx.Config.Rocketchat.DefaultHeader != "" && !noHdr {
			msg += ctx.Config.Rocketchat.DefaultHeader
		}

		msg += fmt.Sprintf(`Requesting review for PR: [%s](%s) (*%s* -> *%s*)`, pr.Title, pr.Url, head, base)

		footer := opts[9].Val.Str
		if footer != "" {
			msg += fmt.Sprintf("\n%s", footer)
		}

		_, err = rctx.PostMessage(&rocketchat.PostMsgRequest{
			Channel: targetChan,
			Text:    msg,
		})

		if err != nil {
			return err
		}
	}

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

func mergePrOpts(c *common.Config) []CmdOpt {
	opts := repoInfoOpts(c)
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

	optional := !c.Rocketchat.Enabled || c.Rocketchat.DefaultNotifyChannel != ""
	opts = append(opts, CmdOpt{
		Spec: CmdOptSpec{
			ArgFlags: []string{"notify"},
			Label:    "Notification channel for rocketchat",
			NoPrompt: optional,
			Optional: optional,
			DefaultStrFunc: func() (string, error) {
				return c.Rocketchat.DefaultNotifyChannel, nil
			},
		},
	})

	return opts
}

func (ctx *CmdCtx) MergePrCommand() error {
	if err := ctx.ValidateConfig(true); err != nil {
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
		Token:  ctx.Config.Gitea.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.Gitea.ToApiUrl(),
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
		if err := repoCtx.DeleteBranch(&gitea.DeleteBranchRequest{
			Branch: pr.Head.Ref,
		}); err != nil {
			return err
		}
	}

	if ctx.Config.Rocketchat.Enabled {
		rctx := rocketchat.Ctx{
			ApiUrl: ctx.Config.Rocketchat.ToApiUrl(),
			UserID: ctx.Config.Rocketchat.UserID,
			Token:  ctx.Config.Rocketchat.Token,
		}

		targetChan := opts[5].Val.Str

		_, err = rctx.PostMessage(&rocketchat.PostMsgRequest{
			Channel: targetChan,
			Text: fmt.Sprintf(`
			[%s](%s) (*%s* -> *%s*) has been merged
		`, pr.Title, pr.Url, pr.Head.Ref, pr.Base.Ref),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

//

func updatePrOpts(config *common.Config) []CmdOpt {
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
	if err := ctx.ValidateConfig(true); err != nil {
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
		Token:  ctx.Config.Gitea.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.Gitea.ToApiUrl(),
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
