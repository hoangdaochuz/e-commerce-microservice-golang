package custom_nats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/tracing"
	"github.com/nats-io/nats.go"
)

type ServerConfig struct {
	ServiceName  string
	OtelEndpoint string
}

type Server struct {
	natsConn        *nats.Conn
	router          *Router
	natsSubject     string
	client          Client
	subcriptions    *nats.Subscription
	ServerConfig    *ServerConfig
	shutdownTracing func()
}

func NewServer(natsConn *nats.Conn, router *Router, natsSubject string, client Client, serverConfig *ServerConfig) *Server {
	return &Server{
		natsConn:     natsConn,
		router:       router,
		natsSubject:  natsSubject,
		client:       client,
		ServerConfig: serverConfig,
		// shutdownTracing: shutdownTracing,
	}
}

func (s *Server) setSubcriptions(subcriptions *nats.Subscription) {
	s.subcriptions = subcriptions
}

func (s *Server) setShutdownTracing(shutdownTracing func()) {
	s.shutdownTracing = shutdownTracing
}
func (s *Server) subcribeNats() (*nats.Subscription, error) {
	handler := func(msg *nats.Msg) {
		var natsRequest Request
		if err := json.Unmarshal(msg.Data, &natsRequest); err != nil {
			logging.GetSugaredLogger().Errorf("fail to unmarshal nats request: %v", err)
			return
		}
		request, err := NatsRequestToHttpRequest(&natsRequest)
		if err != nil {
			logging.GetSugaredLogger().Errorf("fail to change nats request to http request: %v", err)
			return
		}
		// instrument the request
		ctx := tracing.ExtractTraceFromHttpRequest(request)

		response := &Response{
			Headers: http.Header{},
		}
		s.router.ServeHTTP(response, request.WithContext(ctx))

		responseByte, err := json.Marshal(response)
		if err != nil {
			logging.GetSugaredLogger().Errorf("fail to marshal response: %v", err)
			return
		}
		if msg.Reply != "" {
			err := msg.Respond(responseByte)
			if err != nil {
				logging.GetSugaredLogger().Errorf("fail to respond message: %v", err)
				return
			}
		}
	}
	subcriptions, err := s.natsConn.QueueSubscribe(s.natsSubject, "workers", handler)
	if err != nil {
		return nil, fmt.Errorf("fail to subcribe to nats: %w", err)
	}
	s.setSubcriptions(subcriptions)
	return subcriptions, nil
}

func (s *Server) Start() error {

	shutdownTracing, err := tracing.InitializeTraceRegistry(&tracing.TracingConfig{
		ServiceName: s.ServerConfig.ServiceName,
		// Attributes:,
		SamplingRate: 1,
		BatchTimeout: 5 * time.Second,
		BatchMaxSize: 512,
		OtelEndpoint: s.ServerConfig.OtelEndpoint,
	})
	if err != nil {
		logging.GetSugaredLogger().Error(err)
		return err
	}
	s.setShutdownTracing(shutdownTracing)

	s.client.Register(*s.router)
	// subcribe subject
	_, err = s.subcribeNats()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	err := s.subcriptions.Drain()
	if err != nil {
		return err
	}
	s.shutdownTracing()
	logging.GetSugaredLogger().Infof("Tracing has shut down for service %s", s.ServerConfig.ServiceName)
	return nil
}
