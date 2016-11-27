package main

import (
	"flag"
	"os"

	"github.com/admpub/log"
	"github.com/mitool/find2/model"
)

func main() {
	model.CmdOptions.DefineFlag()
	flag.Parse()
	log.Sync()
	if len(os.Args) < 2 {
		//TODO: http server
	}
	model.CmdOptions.Run()
}
