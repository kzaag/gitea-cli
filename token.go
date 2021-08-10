package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"

	"golang.org/x/sys/unix"
)

type TokenRequestBody struct {
	TokenName string `json:"name"`
}

type Credentials struct {
	Username string
	Password string
	TokenRequestBody
}

func (c *Credentials) ToBasicAuth() string {
	resFmt := "Basic %s"
	str := fmt.Sprintf("%s:%s", c.Username, c.Password)
	base64 := base64.StdEncoding.EncodeToString([]byte(str))
	return fmt.Sprintf(resFmt, base64)
}

func (c *Credentials) Validate() error {
	if c.Username == "" || c.Password == "" {
		return fmt.Errorf("invalid username or password")
	}
	return nil
}

const charset = "qwertyuiopasdfghjklzxcvbnm"

var maxInt = big.NewInt(100000)

// may panic
func randStr(l int) string {
	res := make([]byte, l)
	for i := 0; i < l; i++ {
		x, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			panic(err)
		}
		res[i] = charset[x.Int64()%int64(len(charset))]
	}
	return string(res)
}

// fill credentials object from CLI [user input]
// fileds already provided wont be filled
// you dont need to call c.Validate afer this function
func (c *Credentials) FillFromConsole() error {

	fd := unix.Stdin
	const TCSANOW = 0

	if c.Username == "" {
		fmt.Print("Your gitea username: ")
		fmt.Scanln(&c.Username)
	}

	if c.Password == "" {
		s, err := unix.IoctlGetTermios(fd, unix.TCGETS)
		if err != nil {
			return err
		}
		s.Lflag &^= unix.ECHO
		if err := unix.IoctlSetTermios(fd, unix.TCSETS, s); err != nil {
			return err
		}

		fmt.Print("Your gitea password: ")
		fmt.Scanln(&c.Password)
		fmt.Println()

		s, err = unix.IoctlGetTermios(fd, unix.TCGETS)
		if err != nil {
			return err
		}
		s.Lflag |= unix.ECHO
		if err := unix.IoctlSetTermios(fd, unix.TCSETS, s); err != nil {
			return err
		}
	}

	if c.TokenName == "" {
		fmt.Print("Token name (empty for rand): ")
		fmt.Scanln(&c.TokenName)
		if c.TokenName == "" {
			c.TokenName = randStr(8)
		}
	}

	return c.Validate()
}

type Token struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Sha1           string `json:"sha1"`
	TokenLastEight string `json:"token_last_eight"`
}

func (c *Credentials) GetToken() (*Token, error) {
	token := new(Token)
	const m = "POST"
	hdr := make(http.Header)
	hdr.Add("Authorization", c.ToBasicAuth())
	var u = fmt.Sprintf("%s/users/%s/tokens", RepoApiUrl, c.Username)
	return token, GiteaRequest(m, u, &c.TokenRequestBody, token, hdr, 201)
}

func (c *Credentials) DeleteToken(name string) error {
	const m = "DELETE"
	hdr := make(http.Header)
	hdr.Add("Authorization", c.ToBasicAuth())
	var u = fmt.Sprintf("%s/users/%s/tokens/%s", RepoApiUrl, c.Username, name)
	return GiteaRequest(m, u, nil, nil, hdr, 204)
}
