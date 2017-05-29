package main

import (
	"gopkg.in/urfave/cli.v1"

	// Empty imports to have the init-functions called which should
	// register the protocol
	_ "github.com/dedis/cothority/cosi/protocol"
	_ "github.com/dedis/cothority/cosi/service"
	"gopkg.in/dedis/onet.v2/app"
)

func runServer(ctx *cli.Context) {
	// first check the options
	config := ctx.String("config")

	app.RunServer(config)
}
