package gitea

import (
	"fmt"
	"net/http"
)

type ListPRRequest struct {
	State string
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

func (ctx *RepoCtx) ListPR(r *ListPRRequest) ([]PullRequest, error) {
	const m = "GET"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+ctx.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s", ctx.ApiUrl, ctx.Owner, ctx.Repo, r.State)
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
	Opt CreatePullRequestOption
}

func (ctx *RepoCtx) CreatePR(r *CreatePRRequest) (*PullRequest, error) {
	const m = "POST"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+ctx.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/pulls", ctx.ApiUrl, ctx.Owner, ctx.Repo)
	var res = new(PullRequest)
	return res, GiteaRequest(m, u, &r.Opt, res, hdr, 201)
}

type MergePullRequestOption struct {
	ForceMerge bool   `json:"force_merge"`
	Do         string `json:"do"`
}

type MergePRRequest struct {
	Opt   MergePullRequestOption
	Index int
}

func (ctx *RepoCtx) MergePR(r *MergePRRequest) error {
	const m = "POST"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+ctx.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/pulls/%d/merge", ctx.ApiUrl, ctx.Owner, ctx.Repo, r.Index)
	return GiteaRequest(m, u, &r.Opt, nil, hdr, 200)
}
