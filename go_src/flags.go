package main

import (
	"go.wit.com/log"
)

var TERM_EVERYTHING *log.LogFlag
var TERM_EVERYTHING_WARN *log.LogFlag

func init() {
	full := "go.wit.com/gui"
	short := "tree"
	TERM_EVERYTHING = log.NewFlag("TERM.EVERYTHING", false, full, short, "term.everything info")
	TERM_EVERYTHING_WARN = log.NewFlag("TERM.EVERYTHING_WARN", true, full, short, "term.everything warnings")
}
