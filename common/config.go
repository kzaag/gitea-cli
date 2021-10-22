package common

import (
	"fmt"
)

type RocketchatInfo struct {
	ApiVer  string `yaml:"api_ver"`
	BaseUrl string `yaml:"base_url"`
}

func (ri *RocketchatInfo) IsValid() bool {
	if ri.ApiVer == "" || ri.BaseUrl == "" {
		return false
	}
	return true
}

func (c *RocketchatInfo) ToServiceUrl() string {
	return fmt.Sprintf("%s/api/%s", c.BaseUrl, c.ApiVer)
}

type GiteaRepoInfo struct {
	ApiVer  string `yaml:"api_ver"`
	RepoUrl string `yaml:"repo_url"`
}

func (c *GiteaRepoInfo) validationErr(msg string) error {
	return fmt.Errorf("Validate GiteaRepoInfo: %s", msg)
}

func (c *GiteaRepoInfo) Validate() error {
	if c.RepoUrl == "" {
		return c.validationErr("invalid repo_url")
	}
	if c.ApiVer == "" {
		return c.validationErr("invalid api_ver")
	}
	return nil
}

func (c *GiteaRepoInfo) ToRepoApiUrl() string {
	return fmt.Sprintf("%s/api/%s", c.RepoUrl, c.ApiVer)
}

type Config struct {
	TokenSha1 string `yaml:"token_sha1"`
	TokenName string `yaml:"token_name"`
	Username  string `yaml:"username"`

	// Default repository, can be empty
	DefaultRepoName  string `yaml:"default_repo_name"`
	DefaultRepoOwner string `yaml:"default_repo_owner"`

	// this branch will be used as default base for mr's
	DefaultBaseForMr string `yaml:"default_base_for_mr"`

	GiteaRepoInfo  `yaml:"gitea_repo_info"`
	RocketchatInfo `yaml:"rocketchat_info"`

	RocketchatUserID        string `yaml:"rocketchat_user_id"`
	RocketchatToken         string `yaml:"rocketchat_token"`
	RocketchatNotifyChannel string `yaml:"rocketchat_channel"`
}

func (c *Config) validationErr(msg string) error {
	return fmt.Errorf("Validate Config: %s", msg)
}

func (c *Config) Validate() error {
	if c.TokenSha1 == "" {
		return c.validationErr("invalid token_sha1")
	}
	if c.TokenName == "" {
		return c.validationErr("invalid token_name")
	}
	if c.Username == "" {
		return c.validationErr("invalid username")
	}

	if c.RocketchatInfo.IsValid() {
		if c.RocketchatUserID == "" {
			return c.validationErr("invalid rocketchat_user_id")
		}
		if c.RocketchatToken == "" {
			return c.validationErr("invalid rocketchat_token")
		}
		if c.RocketchatNotifyChannel == "" {
			return c.validationErr("invalid rocketchat_channel")
		}
	}

	return c.GiteaRepoInfo.Validate()
}
