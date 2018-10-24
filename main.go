package main

import (
	"github.com/ProtocolONE/p1pay.api/api"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database"
	"log"
)

func main() {
	err, conf := config.NewConfig()

	if err != nil {
		log.Fatalln("unable to get configuration")
	}

	db, err := database.NewConnection(&conf.Database)

	if err != nil {
		log.Fatalf("database connection failed with error: %s\n", err)
	}

	defer db.Close()

	//db.(*mongo).CyrrencyRepository()
	//return

	server, err := api.NewServer(&conf.Jwt, db)

	if err != nil {
		log.Fatalf("server crashed on init with error: %s\n", err)
	}

	err = server.Start()

	if err != nil {
		log.Fatalf("server crashed on start with error: %s\n", err)
	}
}
