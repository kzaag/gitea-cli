package gitea

import (
	"encoding/base64"
	"fmt"
	"gitea-cli/common"
	"net/http"
)

type TokenRequestBody struct {
	TokenName string `json:"name"`
}

type TokenRequest struct {
	Username string
	Password string
	TokenRequestBody
	common.RemoteInfo
}

func (c *TokenRequest) ToBasicAuth() string {
	resFmt := "Basic %s"
	str := fmt.Sprintf("%s:%s", c.Username, c.Password)
	base64 := base64.StdEncoding.EncodeToString([]byte(str))
	return fmt.Sprintf(resFmt, base64)
}

func (c *TokenRequest) Validate() error {
	if c.Username == "" || c.Password == "" {
		return fmt.Errorf("invalid username or password")
	}
	return c.RemoteInfo.Validate()
}

type Token struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Sha1           string `json:"sha1"`
	TokenLastEight string `json:"token_last_eight"`
}

func (c *TokenRequest) GetToken() (*Token, error) {
	token := new(Token)
	const m = "POST"
	hdr := make(http.Header)
	hdr.Add("Authorization", c.ToBasicAuth())
	var u = fmt.Sprintf("%s/users/%s/tokens", c.RemoteInfo.ToApiUrl(), c.Username)
	return token, common.HttpRequest(m, u, &c.TokenRequestBody, token, hdr, 201)
}

func (c *TokenRequest) DeleteToken() error {
	const m = "DELETE"
	hdr := make(http.Header)
	hdr.Add("Authorization", c.ToBasicAuth())
	var u = fmt.Sprintf("%s/users/%s/tokens/%s", c.RemoteInfo.ToApiUrl(), c.Username, c.TokenName)
	return common.HttpRequest(m, u, nil, nil, hdr, 204)
}
