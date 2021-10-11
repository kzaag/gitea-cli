package cmd

import (
	"fmt"
	"gitea-cli/gitea"
	"os"
	"os/exec"
	"strings"
)

func GetMrRepoInfo(ctx *CmdCtx) (repo, owner string, err error) {

	owner = ctx.Config.DefaultRepoOwner
	repo = ctx.Config.DefaultRepoName

	if owner == "" {
		fmt.Print("Provide repo owner: ")
		fmt.Scanln(&owner)
		if owner == "" {
			err = fmt.Errorf("repo owner cannot be empty")
			return
		}
	}

	if repo == "" {
		fmt.Print("Provide repo name: ")
		fmt.Scanln(&repo)
		if repo == "" {
			err = fmt.Errorf("repo name cannot be empty")
			return
		}
	}

	return
}

func MergePR(ctx *CmdCtx) error {
	if ctx.Config == nil {
		return fmt.Errorf("config not found. To create it execute command: config")
	}

	repoCtx := gitea.RepoCtx{}

	var err error
	repoCtx.Repo, repoCtx.Owner, err = GetMrRepoInfo(ctx)
	if err != nil {
		return err
	}
	repoCtx.Token = ctx.Config.TokenSha1
	repoCtx.ApiUrl = ctx.Config.ToRepoApiUrl()
	if err := repoCtx.Validate(); err != nil {
		return err
	}

	title := ""

	if len(os.Args) > 3 {
		title = os.Args[3]
	} else {
		fmt.Print("Provide PR title: ")
		fmt.Scanln(&title)
		if title == "" {
			err = fmt.Errorf("PR title cannot be empty")
			return err
		}
	}

	listRequest := gitea.ListPRRequest{
		State: "open",
	}
	prs, err := repoCtx.ListPR(&listRequest)
	if err != nil {
		return err
	}

	var pr *gitea.PullRequest
	for i := range prs {
		if prs[i].Title == title {
			pr = &prs[i]
			break
		}
	}
	if pr == nil {
		return fmt.Errorf("Couldnt find PR with title: %s", title)
	}

	mergeReq := gitea.MergePRRequest{}
	mergeReq.Opt.Do = "squash"
	mergeReq.Opt.ForceMerge = true
	mergeReq.Index = pr.Number
	if err := repoCtx.MergePR(&mergeReq); err != nil {
		return err
	}

	fmt.Println(pr.Url)

	fmt.Printf("Delete branch %s? [y/n]: ", pr.Head.Ref)
	var db string
	fmt.Scanln(&db)
	if db == "y" {
		delBranchReq := gitea.DeleteBranchRequest{
			Branch: pr.Head.Ref,
		}
		if err := repoCtx.DeleteBranch(&delBranchReq); err != nil {
			return err
		}
	}

	return nil
}

func ListPRs(ctx *CmdCtx) error {

	if ctx.Config == nil {
		return fmt.Errorf("config not found. To create it execute command: config")
	}

	repoCtx := gitea.RepoCtx{}
	var err error
	repoCtx.Repo, repoCtx.Owner, err = GetMrRepoInfo(ctx)
	if err != nil {
		return err
	}
	repoCtx.Token = ctx.Config.TokenSha1
	repoCtx.ApiUrl = ctx.Config.ToRepoApiUrl()
	if err := repoCtx.Validate(); err != nil {
		return err
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

func NewPR(ctx *CmdCtx) error {

	if ctx.Config == nil {
		return fmt.Errorf("config not found. To create it execute command: config")
	}

	repo, owner, err := GetMrRepoInfo(ctx)
	if err != nil {
		return err
	}

	base := ctx.Config.DefaultBaseForMr
	var head = ""

	if base == "" {
		fmt.Print("Provide base branch for mr: ")
		fmt.Scanln(&base)
		if base == "" {
			return fmt.Errorf("base branch cannot be empty")
		}
	}

	fmt.Print("Head [empty to use current head]: ")
	fmt.Scanln(&head)

	if head == "" {
		o, err := exec.Command("git", "branch", "--show-current").Output()
		if err != nil {
			return err
		}
		head = strings.TrimSpace(string(o))
		if head != "" {
			fmt.Printf("Using %s as head\n", head)
		}
	}

	if head == "" {
		return fmt.Errorf("head cannot be empty")
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

func PRCmd(ctx *CmdCtx, argv []string) error {
	cmd := ""
	if len(argv) == 0 {
		cmd = "new"
	} else {
		cmd = argv[0]
	}

	switch cmd {
	case "help":
		fmt.Println(GetHelpStr([]NameWithDesc{
			{
				Name: "help",
				Desc: "prints info",
			},
			{
				Name: "new",
				Desc: "[default]. create new PR",
			},
			{
				Name: "merge",
				Desc: "merge PR",
			},
			{
				Name: "list",
				Desc: "list open PRs",
			},
		}))
	case "new":
		return NewPR(ctx)
	case "merge":
		return MergePR(ctx)
	case "list":
		return ListPRs(ctx)
	default:
		return fmt.Errorf("MR: Command '%s' not found, Use 'mr help' to list available commands", cmd)
	}

	return nil
}
