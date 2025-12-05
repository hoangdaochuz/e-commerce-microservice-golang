package custom_nats

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/nats-io/nats.go"
)

type Server struct {
	natsConn     *nats.Conn
	router       *Router
	natsSubject  string
	client       Client
	subcriptions *nats.Subscription
}

func NewServer(natsConn *nats.Conn, router *Router, natsSubject string, client Client) *Server {
	return &Server{
		natsConn:    natsConn,
		router:      router,
		natsSubject: natsSubject,
		client:      client,
	}
}

func (s *Server) setSubcriptions(subcriptions *nats.Subscription) {
	s.subcriptions = subcriptions
}

func (s *Server) subcribeNats() (*nats.Subscription, error) {
	handler := func(msg *nats.Msg) {
		var natsRequest Request
		json.Unmarshal(msg.Data, &natsRequest)
		request, err := NatsRequestToHttpRequest(&natsRequest)
		if err != nil {
			logging.GetSugaredLogger().Errorf("fail to change nats request to http request: %v", err)
			return
		}

		response := &Response{
			Headers: http.Header{},
		}
		s.router.ServeHTTP(response, request)

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
	s.client.Register(*s.router)
	// subcribe subject
	_, err := s.subcribeNats()
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
	return nil
}
