package rocketchat

import (
	"fmt"
	"gitea-cli/common"
	"net/http"
)

type PostMsgRequest struct {
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text"`
}

type PostMsgResponse struct {
	Success bool `json:"success"`
}

func (ctx *Ctx) PostMessage(req *PostMsgRequest) (*PostMsgResponse, error) {
	res := new(PostMsgResponse)
	hdr := make(http.Header)
	hdr.Add("X-Auth-Token", ctx.Token)
	hdr.Add("X-User-Id", ctx.UserID)
	var u = fmt.Sprintf("%s/chat.postMessage", ctx.ApiUrl)
	if err := common.HttpRequest("POST", u, req, res, hdr, 200); err != nil {
		return nil, err
	}
	if !res.Success {
		return nil, fmt.Errorf("Request wasn't successful")
	}
	return res, nil
}
