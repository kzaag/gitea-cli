package rocketchat

import (
	"fmt"
	"gitea-cli/common"
)

type LoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status string `json:"status"`
	Data   struct {
		AuthToken string `json:"authToken"`
		UserID    string `json:"userId"`
	} `json:"data"`
	Message string `json:"message,omitempty"`
}

func Login(r *common.RemoteInfo, req *LoginRequest) (*LoginResponse, error) {
	res := new(LoginResponse)
	var u = fmt.Sprintf("%s/login", r.ToApiUrl())
	if err := common.HttpRequest("POST", u, req, res, nil, 200); err != nil {
		return nil, err
	}
	if res.Status != "success" {
		return nil, fmt.Errorf("unexpected status: %s", res.Status)
	}
	return res, nil
}
