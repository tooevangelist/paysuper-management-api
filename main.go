package main

import (
	"github.com/ProtocolONE/p1pay.api/api"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database"
	"go.uber.org/zap"
	"log"
)

func main() {
	err, conf := config.NewConfig()

	if err != nil {
		log.Fatalln(err)
	}

	db, err := database.NewConnection(&conf.Database)

	if err != nil {
		log.Fatalf("database connection failed with error: %s\n", err)
	}

	defer db.Close()

	logger, err := zap.NewProduction()

	if err != nil {
		log.Fatalf("Application logger initialization failed with error: %s\n", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			return
		}
	}()
	sugar := logger.Sugar()

	server, err := api.NewServer(&conf.Jwt, db, sugar)

	if err != nil {
		log.Fatalf("server crashed on init with error: %s\n", err)
	}

	err = server.Start()

	if err != nil {
		log.Fatalf("server crashed on start with error: %s\n", err)
	}
}
