package gitea

import "fmt"

type RepoCtx struct {
	Owner  string
	Repo   string
	Token  string
	ApiUrl string
}

func (rctx *RepoCtx) Validate() error {
	if rctx.Owner == "" {
		return fmt.Errorf("validate RepoCtx: no repository owner provided")
	}
	if rctx.Repo == "" {
		return fmt.Errorf("validate RepoCtx: no repository name provided")
	}
	if rctx.Token == "" {
		return fmt.Errorf("validate RepoCtx: no auth token provided")
	}
	if rctx.ApiUrl == "" {
		return fmt.Errorf("validate RepoCtx: no Api url provided")
	}
	return nil
}
