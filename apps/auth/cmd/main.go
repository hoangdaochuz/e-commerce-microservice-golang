package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	authService_api "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/api/auth"
	auth_handler "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/handler/auth"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/nats-io/nats.go"

	// Import để trigger dependency registration
	_ "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/handler/session"
	_ "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/services/auth"
	_ "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/cache"
	_ "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/configs/redis"
	_ "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/redis"
)

const (
	NATS_URL_KEY          = "nats_auth.nats_url"
	NATS_APP_USERNAME_KEY = "nats_auth.nats_apps.0.username"
	NATS_APP_PASSWORD_KEY = "nats_auth.nats_apps.0.password"
)

func main() {
	config, err := configs.Load()
	if err != nil {
		log.Fatal("fail to load config")
	}

	natsConfig := custom_nats.GetNatsConfig()

	// natsUrl := config.NatsAuth.NATSUrl
	// natsUsername := config.NatsAuth.NATSApps[0].Username
	// natsPass := config.NatsAuth.NATSApps[0].Password
	// Connect to nats
	natsConn, err := nats.Connect(natsConfig.NatsUrl, nats.UserInfo(natsConfig.NatsAppAccUserName, natsConfig.NatsAppAccountPassword))
	if err != nil {
		log.Fatal("fail to connect to nats: %w", err)
	}

	chiRouter := chi.NewRouter()
	router := custom_nats.NewRouter(chiRouter)

	var authServiceApp authService_api.AuthenticateService

	err = di.Resolve(func(authServiceAppImplement *auth_handler.AuthServiceApp) {
		authServiceApp = authServiceAppImplement
	})
	if err != nil {
		log.Fatal("fail to get auth service app")
	}

	authServiceProxy := authService_api.NewAuthenticateServiceProxy(authServiceApp)
	authServiceClient := authService_api.NewAuthenticateServiceRouter(authServiceProxy)

	server := custom_nats.NewServer(natsConn, router, authService_api.NATS_SUBJECT, authServiceClient, &custom_nats.ServerConfig{
		ServiceName:  "auth",
		OtelEndpoint: config.GeneralConfig.OTLP_Endpoint,
	})

	err = server.Start()
	if err != nil {
		_ = server.Stop()
		log.Fatal("fail to start auth service app")
	}
	shutdowSign := make(chan os.Signal, 1)
	signal.Notify(shutdowSign, syscall.SIGINT, syscall.SIGTERM)
	<-shutdowSign
	logging.GetSugaredLogger().Infof("Shutting down authenticate service server")
	err = server.Stop()
	if err != nil {
		logging.GetSugaredLogger().Fatalf("fail to shut down authenticate service server: %v", err)
	}
	logging.GetSugaredLogger().Infof("Shut down authenticate service server peacefully successfully")
}
