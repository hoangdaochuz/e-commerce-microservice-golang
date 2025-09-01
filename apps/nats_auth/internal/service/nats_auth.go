package nats_auth_service

import (
	"fmt"
	"log"

	jwt "github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
)

type Server struct {
	natsConn    *nats.Conn
	authSubject string
	subcription *nats.Subscription
}

func NewServer(natsConn *nats.Conn, authSubject string) *Server {
	return &Server{
		natsConn:    natsConn,
		authSubject: authSubject,
	}
}

func (s *Server) decryptPayloadMessage(msg *nats.Msg) (*jwt.AuthorizationRequestClaims, error) {
	// Load xkey private key
	// xkeyPrivate := "SXAOY7ATGBMA5CCJ2XWD3XFH7K3TDFAJDHLTMVSCRJCHCESOAEHEGWKPMI"
	// kp, err := nkeys.FromSeed([]byte(xkeyPrivate))
	// if err != nil {
	// 	return nil, err
	// }
	fmt.Println("Wtf")
	reqClaims, err := jwt.DecodeAuthorizationRequestClaims(string(msg.Data))
	log.Default().Print("err: ", err)
	if err != nil {
		return nil, err
	}
	fmt.Println("Name", reqClaims.Name)
	return reqClaims, nil
}

func (s *Server) Handler(msg *nats.Msg) {
	_, err := s.decryptPayloadMessage(msg)
	if err != nil {
		fmt.Errorf("fail to decrypt msg: %w", err)
	}
	fmt.Println("Hello this is nats auth handler")

}

func (s *Server) Listen() error {
	sub, err := s.natsConn.Subscribe(s.authSubject, s.Handler)
	if err != nil {
		return err
	}
	s.subcription = sub
	return nil
}

func (s *Server) Stop() error {
	if s.subcription != nil {
		err := s.subcription.Unsubscribe()
		if err != nil {
			return err
		}
	}

	if s.natsConn != nil {
		err := s.natsConn.Drain()
		if err != nil {
			return err
		}
	}
	return nil
}
