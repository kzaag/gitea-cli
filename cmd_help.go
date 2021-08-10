package main

import "fmt"

func HelpHandler(ctx *AppCtx, argv []string) error {
	nd := make([]NameDesc, len(ctx.Commands))
	i := 0
	for x := range ctx.Commands {
		nd[i] = NameDesc{
			Name: x,
			Desc: ctx.Commands[x].Desc,
		}
		i++
	}
	fmt.Println(GetHelpStr(nd))
	return nil
}
