package cmd

import (
	"fmt"
	"strings"
)

type NameWithDesc struct {
	Name string
	Desc string
}

func GetHelpStr(commands []NameWithDesc) string {
	sb := strings.Builder{}
	for i := range commands {
		sb.WriteString(fmt.Sprintf("\t%s: %s\n", commands[i].Name, commands[i].Desc))
	}
	return sb.String()
}

func GetArg()
