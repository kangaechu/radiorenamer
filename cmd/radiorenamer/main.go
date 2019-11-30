package main

import (
	"flag"
	"fmt"
	"github.com/kangaechu/radiorenamer"
	"os"
)

var version string
var revision string

func main() {
	flag.Parse()
	if flag.Arg(0) == "version" {
		fmt.Printf(os.Args[0]+": %s-%s\n", version, revision)
		os.Exit(0)
	}

	radiorenamer.Run(flag.Arg(0))
}
