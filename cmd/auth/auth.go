package main

import (
	"github.com/bratushkadan/floral/api/auth"
)

const webPort = "80"

func main() {
	conf := auth.NewAuthServerConfig(48612)
	auth.RunServer(conf)
}
