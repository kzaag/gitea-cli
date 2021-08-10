package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

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

type PullRequest struct {
	ID int64 `yaml:"id"`
}

func CreatePR(r *CreatePRRequest) (*PullRequest, error) {
	const m = "POST"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+r.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/pulls", r.ApiUrl, r.Owner, r.Repo)
	var res = new(PullRequest)
	return res, GiteaRequest(m, u, &r.Opt, r, hdr, 201)
}

func MrHandler(ctx *AppCtx, argv []string) error {
	if ctx.Config == nil {
		return fmt.Errorf("config not found. To create it execute command: config")
	}

	repoOwner := ctx.Config.DefaultRepoOwner
	repoName := ctx.Config.DefaultRepoName
	base := ctx.Config.DefaultBaseForMr
	var head = ""

	if repoOwner == "" {
		fmt.Print("Provide repo owner: ")
		fmt.Scanln(&repoOwner)
		if repoOwner == "" {
			return fmt.Errorf("repo owner cannot be empty")
		}
	}

	if repoName == "" {
		fmt.Print("Provide repo name: ")
		fmt.Scanln(&repoName)
		if repoName == "" {
			return fmt.Errorf("repo name cannot be empty")
		}
	}

	if repoName == "" {
		fmt.Print("Provide repo name: ")
		fmt.Scanln(&repoName)
		if repoName == "" {
			return fmt.Errorf("repo name cannot be empty")
		}
	}

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
		Owner:  repoOwner,
		Repo:   repoName,
		ApiUrl: ctx.Config.ToRepoApiUrl(),
	}

	pr, err := CreatePR(&req)
	if err != nil {
		return err
	}

	fmt.Printf("Credated MR @ %s/%s/%s/pulls/%d\n",
		ctx.Config.ToRepoApiUrl(),
		repoOwner, repoName,
		pr.ID)

	return nil
}
