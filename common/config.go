package common

import (
	"fmt"
)

type RemoteInfo struct {
	ApiVer  string `yaml:"api_ver"`
	BaseUrl string `yaml:"base_url"`
}

func (c *RemoteInfo) ToApiUrl() string {
	return fmt.Sprintf("%s/api/%s", c.BaseUrl, c.ApiVer)
}

func (c *RemoteInfo) validationErr(msg string) error {
	return fmt.Errorf("Validate RemoteInfo: %s", msg)
}

func (c *RemoteInfo) Validate() error {
	if c.BaseUrl == "" {
		return c.validationErr("invalid base_url")
	}
	if c.ApiVer == "" {
		return c.validationErr("invalid api_ver")
	}
	return nil
}

type GiteaConfig struct {
	TokenSha1 string `yaml:"token_sha1"`
	TokenName string `yaml:"token_name"`

	// Default repository, can be empty
	DefaultRepoName  string `yaml:"default_repo_name"`
	DefaultRepoOwner string `yaml:"default_repo_owner"`

	// this branch will be used as default base for mr's
	DefaultBaseForPR string `yaml:"default_base_for_pr"`

	RemoteInfo `yaml:"remote_info"`
}

func (c *GiteaConfig) validationErr(msg string) error {
	return fmt.Errorf("Validate GiteaConfig: %s", msg)
}

func (c *GiteaConfig) Validate(cred bool) error {

	if cred {
		if c.TokenSha1 == "" {
			return c.validationErr("invalid token_sha1")
		}
		if c.TokenName == "" {
			return c.validationErr("invalid token_name")
		}
	}

	return c.RemoteInfo.Validate()
}

type Rocketchat struct {
	Enabled bool `yaml:"enabled"`

	RemoteInfo `yaml:"remote_info"`

	UserID               string `yaml:"user_id"`
	Token                string `yaml:"token"`
	DefaultNotifyChannel string `yaml:"default_notify_channel"`
	DefaultHeader        string `yaml:"default_header"`
}

func (c *Rocketchat) validationErr(msg string) error {
	return fmt.Errorf("Validate Rocketchat: %s", msg)
}

func (c *Rocketchat) Validate(cred bool) error {
	if cred {
		if c.UserID == "" {
			return c.validationErr("invalid user_id")
		}
		if c.Token == "" {
			return c.validationErr("invalid token")
		}
	}
	return c.RemoteInfo.Validate()
}

type Config struct {
	Gitea      GiteaConfig
	Rocketchat Rocketchat
}

func (c *Config) validationErr(msg string) error {
	return fmt.Errorf("Validate Config: %s", msg)
}

func (c *Config) Validate(cred bool) error {
	if err := c.Gitea.Validate(cred); err != nil {
		return err
	}
	if c.Rocketchat.Enabled {
		return c.Rocketchat.Validate(cred)
	}
	return nil
}
