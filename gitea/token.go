package gitea

import (
	"encoding/base64"
	"fmt"
	"gitea-cli/config"
	"net/http"
)

type TokenRequestBody struct {
	TokenName string `json:"name"`
}

type TokenRequest struct {
	Username string
	Password string
	TokenRequestBody
	config.RepoInfo
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
	return c.RepoInfo.Validate()
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
	var u = fmt.Sprintf("%s/users/%s/tokens", c.RepoInfo.ToRepoApiUrl(), c.Username)
	return token, GiteaRequest(m, u, &c.TokenRequestBody, token, hdr, 201)
}

func (c *TokenRequest) DeleteToken(name string) error {
	const m = "DELETE"
	hdr := make(http.Header)
	hdr.Add("Authorization", c.ToBasicAuth())
	var u = fmt.Sprintf("%s/users/%s/tokens/%s", c.RepoInfo.ToRepoApiUrl(), c.Username, name)
	return GiteaRequest(m, u, nil, nil, hdr, 204)
}
