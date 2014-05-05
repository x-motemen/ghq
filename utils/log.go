package utils

import (
	"fmt"
	"os"

	"github.com/daviddengcn/go-colortext"
)

func Log(prefix string, message string) {
	if prefix == "git" || prefix == "hg" || prefix == "skip" {
		ct.ChangeColor(ct.White, false, ct.None, false)
	} else if prefix == "open" || prefix == "exists" {
		ct.ChangeColor(ct.Yellow, false, ct.None, false)
	} else if prefix == "authorized" || prefix == "skip" {
		ct.ChangeColor(ct.Blue, false, ct.None, false)
	} else if prefix == "error" {
		ct.ChangeColor(ct.Red, false, ct.None, false)
	} else {
		ct.ChangeColor(ct.Green, false, ct.None, false)
	}
	fmt.Printf("%10s", prefix)
	ct.ResetColor()
	fmt.Printf(" %s\n", message)
}

func ErrorIf(err error) bool {
	if err != nil {
		Log("error", err.Error())
		return true
	} else {
		return false
	}
}

func DieIf(err error) {
	if err != nil {
		Log("error", err.Error())
		os.Exit(1)
	}
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}
