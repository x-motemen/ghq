package main

import (
	"fmt"

	"github.com/daviddengcn/go-colortext"
)

func logInfo(prefix string, message string) {
	if prefix == "git" {
		ct.ChangeColor(ct.White, false, ct.None, false)
	} else if prefix == "error" {
		ct.ChangeColor(ct.Red, false, ct.None, false)
	} else {
		ct.ChangeColor(ct.Green, true, ct.None, false)
	}
	fmt.Printf("%8s", prefix)
	ct.ResetColor()
	fmt.Printf(" %s\n", message)
}
