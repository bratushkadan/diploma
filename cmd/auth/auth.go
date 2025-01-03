package main

import (
	"log"

	"github.com/bratushkadan/floral/api/auth"
)

const webPort = 48612

func main() {
	conf := auth.NewAuthServerConfig(webPort)
	err := auth.RunServer(conf)
	if err != nil {
		log.Fatal(err)
	}
}
