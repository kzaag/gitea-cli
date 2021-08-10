package main

import (
	"fmt"
	"strings"
)

type NameDesc struct {
	Name string
	Desc string
}

func GetHelpStr(commands []NameDesc) string {
	sb := strings.Builder{}
	sb.WriteString("Commands:")
	for i := range commands {
		sb.WriteString(fmt.Sprintf("\n\t%s: %s", commands[i].Name, commands[i].Desc))
	}
	return sb.String()
}
