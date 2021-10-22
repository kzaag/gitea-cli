package rocketchat

import (
	"fmt"
	"gitea-cli/common"
	"net/http"
)

type LoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status string `json:"status"`
	Data   struct {
		AuthToken string `json:"authToken"`
	} `json:"data"`
	Message string `json:"message,omitempty"`
}

func Login(r *common.RocketchatInfo, req *LoginRequest) (*LoginResponse, error) {
	res := new(LoginResponse)
	hdr := make(http.Header)
	var u = fmt.Sprintf("%s/login", r.ToServiceUrl())
	if err := common.HttpRequest("POST", u, req, res, hdr, 200); err != nil {
		return nil, err
	}
	if res.Status != "success" {
		return nil, fmt.Errorf("unexpected status: %s", res.Status)
	}
	return res, nil
}
