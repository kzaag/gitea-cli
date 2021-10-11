package cmd

import "fmt"

func HelpCmd(ctx *CmdCtx, argv []string) error {
	nd := make([]NameWithDesc, len(ctx.Commands))
	i := 0
	for x := range ctx.Commands {
		nd[i] = NameWithDesc{
			Name: x,
			Desc: ctx.Commands[x].Desc,
		}
		i++
	}
	fmt.Println(GetHelpStr(nd))
	return nil
}
