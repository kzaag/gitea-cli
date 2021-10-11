package gitea

import (
	"fmt"
	"net/http"
)

type DeleteBranchRequest struct {
	Branch string
}

func (ctx *RepoCtx) DeleteBranch(r *DeleteBranchRequest) error {
	const m = "DELETE"
	hdr := make(http.Header)
	hdr.Add("Authorization", "token "+ctx.Token)
	var u = fmt.Sprintf("%s/repos/%s/%s/branches/%s", ctx.ApiUrl, ctx.Owner, ctx.Repo, r.Branch)
	return GiteaRequest(m, u, nil, nil, hdr, 204)
}
