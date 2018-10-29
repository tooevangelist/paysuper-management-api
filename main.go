package main

import (
	"flag"
	"github.com/ProtocolONE/p1pay.api/api"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database"
	"github.com/globalsign/mgo"
	"go.uber.org/zap"
	"log"
)

func main() {
	migration := flag.String("migration", "", "run database migrations with specified direction")
	flag.Parse()

	err, conf := config.NewConfig()

	if err != nil {
		log.Fatalln(err)
	}

	db, err := database.NewConnection(&conf.Database)

	if err != nil {
		log.Fatalf("database connection failed with error: %s\n", err)
	}

	defer db.Close()

	if *migration != "" {
		err := database.Migrate(db.Database().(*mgo.Database), *migration)

		if err != nil {
			log.Fatalf("database migration failed with error: %s\n", err)
		}

		return
	}

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
