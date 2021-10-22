package rocketchat

import "fmt"

type Ctx struct {
	ApiUrl string
	UserID string
	Token  string
}

func (ctx *Ctx) Validate() error {
	if ctx.ApiUrl == "" {
		return fmt.Errorf("validate rocketchat.Ctx: invalid ApiUrl")
	}
	if ctx.UserID == "" {
		return fmt.Errorf("validate rocketchat.Ctx: invalid UserID")
	}
	if ctx.Token == "" {
		return fmt.Errorf("validate rocketchat.Ctx: invalid Token")
	}
	return nil
}
