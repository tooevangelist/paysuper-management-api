package main

import (
	"context"
	"github.com/ProtocolONE/p1pay.api/api"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database"
	"github.com/ProtocolONE/payone-repository/pkg/constant"
	"github.com/micro/go-micro"
	_ "github.com/micro/go-plugins/broker/rabbitmq"
	_ "github.com/micro/go-plugins/transport/grpc"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// @title Protocol One payment solution swagger documentation
// @version 1.0
// @description This is a Protocol One payment solution service.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host p1payapi.tst.protocol.one
func main() {
	//migration := flag.String("migration", "", "run database migrations with specified direction")
	//flag.Parse()

	err, conf := config.NewConfig()

	if err != nil {
		log.Fatalln(err)
	}

	db, err := database.NewConnection(&conf.Database)

	if err != nil {
		log.Fatalf("database connection failed with error: %s\n", err)
	}

	defer db.Close()

	/*if *migration != "" {
		err := database.Migrate(db.Database().(*mgo.Database), *migration)

		if err != nil {
			log.Fatalf("database migration failed with error: %s\n", err)
		}

		return
	}*/

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

	geoDbReader, err := geoip2.Open(conf.GeoIP.DBPath)

	if err != nil {
		log.Fatalf("geo ip database load failed with error: %s\n", err)
	}
	defer func() {
		if err := geoDbReader.Close(); err != nil {
			return
		}
	}()

	service, publisher, cancelFunc := InitService()

	go func() {
		if err := service.Run(); err != nil {
			return
		}
	}()

	sInit := &api.ServerInitParams{
		Config:                  &conf.Jwt,
		Database:                db,
		Logger:                  sugar,
		GeoDbReader:             geoDbReader,
		PaymentSystemConfig:     conf.PaymentSystemConfig.Config,
		PSPAccountingCurrencyA3: conf.PSPAccountingCurrencyA3,
		HttpScheme:              conf.HttpScheme,
		Publisher:               publisher,
	}

	server, err := api.NewServer(sInit)

	if err != nil {
		log.Fatalf("server crashed on init with error: %s\n", err)
	}

	err = server.Start()

	if err != nil {
		log.Fatalf("server crashed on start with error: %s\n", err)
	}

	handleOsSignals(cancelFunc)
}

func InitService() (micro.Service, micro.Publisher, context.CancelFunc) {
	serviceContext, serviceCancel := context.WithCancel(context.Background())

	service := micro.NewService(
		micro.Name("go.p1.payone.api"),
		micro.Version(constant.PayOneMicroserviceVersion),
		micro.Context(serviceContext),
	)
	service.Init()
	publisher := micro.NewPublisher(constant.PayOneTopicNotifyPaymentName, service.Client())

	return service, publisher, serviceCancel
}

func handleOsSignals(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	exitChan := make(chan int)

	go func() {
		for {
			s := <-signalChan
			switch s {
			case os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT:
				cancelFunc()
				exitChan <- 0
			}
		}
	}()

	code := <-exitChan
	os.Exit(code)
}
