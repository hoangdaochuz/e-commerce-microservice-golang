package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	nats_auth_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/nats_auth/internal/service"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	"github.com/nats-io/nats.go"
)

func main() {
	config, err := configs.Load()
	xkeyPrivate := config.NatsAuth.XKeyPrivate
	if err != nil {
		log.Fatal("failed to load configuration: %w", err)
	}
	var natAuthApp configs.NATSApp
	for _, app := range config.NatsAuth.NATSApps {
		if app.Account == "AUTH" {
			natAuthApp = app
		}
	}
	natsConn, err := nats.Connect(config.NatsAuth.NATSUrl, nats.UserInfo(natAuthApp.Username, natAuthApp.Password))
	if err != nil {
		log.Fatal("[NATS Auth] failed to connect to nats: %w", err)
	}

	natsApps := config.NatsAuth.NATSApps
	issuerPrivate := config.NatsAuth.IssuerPrivate

	server, err := nats_auth_service.NewServer(natsConn, config.NatsAuth.AuthCallOutSubject, xkeyPrivate, natsApps, issuerPrivate)
	if err != nil {
		log.Fatal("failed to create nats auth server: %w", err)
	}
	err = server.Listen()
	if err != nil {
		panic("Fail to listen to nats auth subject")
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Default().Println("Shuttingg down nats auth")
	err = server.Stop()
	if err != nil {
		log.Fatal(fmt.Errorf("fail to shutting down nats auth server %w", err))
	}
}
