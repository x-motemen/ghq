package utils

import (
	"fmt"

	"github.com/daviddengcn/go-colortext"
)

func Log(prefix string, message string) {
	if prefix == "git" {
		ct.ChangeColor(ct.White, false, ct.None, false)
	} else if prefix == "error" {
		ct.ChangeColor(ct.Red, false, ct.None, false)
	} else if prefix == "open" {
		ct.ChangeColor(ct.Yellow, true, ct.None, false)
	} else if prefix == "authorized" || prefix == "skip" {
		ct.ChangeColor(ct.Blue, false, ct.None, false)
	} else {
		ct.ChangeColor(ct.Green, true, ct.None, false)
	}
	fmt.Printf("%10s", prefix)
	ct.ResetColor()
	fmt.Printf(" %s\n", message)
}
