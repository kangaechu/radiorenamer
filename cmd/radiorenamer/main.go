package main

import (
	"flag"
	"github.com/kangaechu/radiorenamer"
)

func main() {
	flag.Parse()
	radiorenamer.Run(flag.Arg(0))
}
