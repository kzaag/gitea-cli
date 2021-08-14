package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type DeleteBranchRequest struct {
	Owner  string
	Repo   string
	Token  string
	ApiUrl string
	Branch string
}

func GiteaDeleteBranch(r *DeleteBranchRequest) error {
	const m = "DELETE"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+r.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/branches/%s", r.ApiUrl, r.Owner, r.Repo, r.Branch)
	return GiteaRequest(m, u, nil, nil, hdr, 204)
}

type ListPRRequest struct {
	Owner  string
	Repo   string
	State  string
	Token  string
	ApiUrl string
}

type PRBranchInfo struct {
	Ref string `json:"ref"`
}

type PullRequest struct {
	Url   string `json:"url"`
	Title string `json:"title"`
	ID    int    `json:"id"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"`
	Base, Head PRBranchInfo
	Number     int `json:"number"`
}

func GiteaListPR(r *ListPRRequest) ([]PullRequest, error) {
	const m = "GET"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+r.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s", r.ApiUrl, r.Owner, r.Repo, r.State)
	var res []PullRequest
	return res, GiteaRequest(m, u, nil, &res, hdr, 200)
}

type CreatePullRequestOption struct {
	//Assignee  string   `json:"assignee"`
	//Assignees []string `json:"assignees"`
	Base string `json:"base"`
	//Body string `json:"body"`
	//DueDate   time.Time `json:"due_date"`
	Head string `json:"head"`
	//Labels    []string  `json:"labels"`
	//Milestone int       `json:"milestone"`
	Title string `json:"title"`
}

type CreatePRRequest struct {
	Opt    CreatePullRequestOption
	Token  string
	Owner  string
	Repo   string
	ApiUrl string
}

func GiteaCreatePR(r *CreatePRRequest) (*PullRequest, error) {
	const m = "POST"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+r.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/pulls", r.ApiUrl, r.Owner, r.Repo)
	var res = new(PullRequest)
	return res, GiteaRequest(m, u, &r.Opt, res, hdr, 201)
}

type MergePullRequestOption struct {
	ForceMerge bool   `json:"force_merge"`
	Do         string `json:"do"`
}

type MergePRRequest struct {
	Opt    MergePullRequestOption
	Owner  string
	Repo   string
	Index  int
	ApiUrl string
	Token  string
}

func GetMrRepoInfo(ctx *AppCtx) (repo, owner string, err error) {

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

func GiteaMergePR(r *MergePRRequest) error {
	const m = "POST"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+r.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/pulls/%d/merge", r.ApiUrl, r.Owner, r.Repo, r.Index)
	return GiteaRequest(m, u, &r.Opt, nil, hdr, 200)
}

func MergePR(ctx *AppCtx) error {
	if ctx.Config == nil {
		return fmt.Errorf("config not found. To create it execute command: config")
	}

	req := MergePRRequest{}
	req.Opt.Do = "squash"
	req.Opt.ForceMerge = true

	var err error
	req.Repo, req.Owner, err = GetMrRepoInfo(ctx)
	if err != nil {
		return err
	}

	req.Token = ctx.Config.TokenSha1
	req.ApiUrl = ctx.Config.ToRepoApiUrl()

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

	listRequest := ListPRRequest{
		Owner:  req.Owner,
		Repo:   req.Repo,
		State:  "open",
		Token:  req.Token,
		ApiUrl: req.ApiUrl,
	}
	prs, err := GiteaListPR(&listRequest)
	if err != nil {
		return err
	}

	var pr *PullRequest
	for i := range prs {
		if prs[i].Title == title {
			pr = &prs[i]
			break
		}
	}

	if pr == nil {
		return fmt.Errorf("Couldnt find PR with title: %s", title)
	}

	req.Index = pr.Number

	if err := GiteaMergePR(&req); err != nil {
		return err
	}

	fmt.Println(pr.Url)

	fmt.Printf("Delete branch %s? [y/n]: ", pr.Head.Ref)
	var db string
	fmt.Scanln(&db)
	if db == "y" {
		dbr := DeleteBranchRequest{
			Owner:  req.Owner,
			Repo:   req.Repo,
			Token:  req.Token,
			ApiUrl: req.ApiUrl,
			Branch: pr.Head.Ref,
		}
		if err := GiteaDeleteBranch(&dbr); err != nil {
			return err
		}
	}

	return nil
}

func ListPRs(ctx *AppCtx) error {

	if ctx.Config == nil {
		return fmt.Errorf("config not found. To create it execute command: config")
	}

	req := ListPRRequest{}
	req.State = "open"
	var err error
	req.Repo, req.Owner, err = GetMrRepoInfo(ctx)
	if err != nil {
		return err
	}

	req.Token = ctx.Config.TokenSha1
	req.ApiUrl = ctx.Config.ToRepoApiUrl()

	prs, err := GiteaListPR(&req)
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

func NewPR(ctx *AppCtx) error {

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

	var req = CreatePRRequest{
		Opt: CreatePullRequestOption{
			Base:  base,
			Head:  head,
			Title: head,
		},
		Token:  ctx.Config.TokenSha1,
		Owner:  owner,
		Repo:   repo,
		ApiUrl: ctx.Config.ToRepoApiUrl(),
	}

	pr, err := GiteaCreatePR(&req)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", pr.Url)

	return nil
}

func PRHandler(ctx *AppCtx, argv []string) error {
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
		return fmt.Errorf("MR: Command '%s' not found, Use 'mr help' to list available commands", argv[1])
	}

	return nil
}
