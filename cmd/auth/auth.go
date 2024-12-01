package main

import (
	"log"

	"github.com/bratushkadan/floral/api/auth"
)

const webPort = "80"

func main() {
	conf := auth.NewAuthServerConfig(48612)
	err := auth.RunServer(conf)
	if err != nil {
		log.Fatal(err)
	}
}
