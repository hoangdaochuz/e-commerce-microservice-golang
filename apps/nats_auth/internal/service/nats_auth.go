package nats_auth_service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	jwt "github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

type Server struct {
	natsConn      *nats.Conn
	authSubject   string
	subcription   *nats.Subscription
	keyPairs      nkeys.KeyPair
	issuerPrivKey string
	natsApps      []configs.NATSApp
}

func NewServer(natsConn *nats.Conn, authSubject string, xkeyPrivate string, natsApps []configs.NATSApp, issuerPrivate string) (*Server, error) {
	keyPairs, err := nkeys.FromSeed([]byte(strings.TrimSpace(xkeyPrivate)))
	if err != nil {
		return nil, err
	}
	return &Server{
		natsConn:      natsConn,
		authSubject:   authSubject,
		keyPairs:      keyPairs,
		natsApps:      natsApps,
		issuerPrivKey: issuerPrivate,
	}, nil
}

func (s *Server) decryptPayloadMessage(msg *nats.Msg) (*jwt.AuthorizationRequestClaims, error) {
	encrypedData := msg.Data
	// senderPubKey := msg.Header.Get("Nats-Server-Xkey")
	var senderPubKey string
	for k, v := range msg.Header {
		if k == "Nats-Server-Xkey" {
			senderPubKey = v[0]
		}
	}
	decryptedData, err := s.keyPairs.Open(encrypedData, senderPubKey)
	if err != nil {
		return nil, err
	}
	fmt.Println("Decrypted data: ", string(decryptedData))

	reqClaims, err := jwt.DecodeAuthorizationRequestClaims(string(decryptedData))
	if err != nil {
		return nil, err
	}
	fmt.Println("Name", reqClaims.AuthorizationRequest.ConnectOptions.Username)
	return reqClaims, nil
}

func (s *Server) encryptedResponse(response []byte, claims *jwt.AuthorizationRequestClaims) ([]byte, error) {
	ontimePublicXKey := claims.AuthorizationRequest.Server.XKey
	fmt.Println("ontimePublicXKey: ", ontimePublicXKey)
	encryptedResponse, err := s.keyPairs.Seal(response, ontimePublicXKey)
	if err != nil {
		return nil, err
	}
	return encryptedResponse, nil
}

func (s *Server) buildCommonResponseAuthorizationClaims(reqClaims *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	fmt.Println("issuerKey: ", s.issuerPrivKey)
	issuerSeed, err := nkeys.FromSeed([]byte(s.issuerPrivKey))
	if err != nil {
		return nil, err
	}
	issuerPubKey, err := issuerSeed.PublicKey()
	if err != nil {
		return nil, err
	}
	userNKey := reqClaims.AuthorizationRequest.UserNkey
	responseClaims := jwt.NewAuthorizationResponseClaims(userNKey)
	responseClaims.Issuer = issuerPubKey
	responseClaims.IssuedAt = time.Now().Unix()
	responseClaims.Audience = reqClaims.AuthorizationRequest.Server.ID
	return responseClaims, nil
}

func (s *Server) sendUnauthorizedResponse(msg *nats.Msg, reqClaims *jwt.AuthorizationRequestClaims) error {
	issuerSeed, err := nkeys.FromSeed([]byte(s.issuerPrivKey))
	if err != nil {
		return err
	}
	responseClaims, err := s.buildCommonResponseAuthorizationClaims(reqClaims)
	if err != nil {
		return err
	}
	responseClaims.Error = "Mismatch username or password of client connection"

	responseToken, err := responseClaims.Encode(issuerSeed)
	if err != nil {
		return err
	}

	responseMsg, err := s.encryptedResponse([]byte(responseToken), reqClaims)
	if err != nil {
		return err
	}

	err = msg.Respond(responseMsg)
	if err != nil {
		return err
	}
	fmt.Println("Sent unauthorized response")
	return nil
}

func (s *Server) sendAuthorizedResponse(msg *nats.Msg, reqClaims *jwt.AuthorizationRequestClaims) error {
	issuerSeed, err := nkeys.FromSeed([]byte(s.issuerPrivKey))
	if err != nil {
		return err
	}
	issuerPubKey, err := issuerSeed.PublicKey()
	if err != nil {
		return err
	}

	responseClaims, err := s.buildCommonResponseAuthorizationClaims(reqClaims)
	if err != nil {
		return err
	}
	userClaims := jwt.NewUserClaims(reqClaims.AuthorizationRequest.UserNkey)
	userClaims.Name = reqClaims.AuthorizationRequest.ConnectOptions.Username
	userClaims.Issuer = issuerPubKey
	userClaims.IssuedAt = time.Now().Unix()
	userClaims.Audience = "APP"

	if userClaims.Name == "admin" {
		userClaims.Permissions = jwt.Permissions{
			Pub: jwt.Permission{
				Allow: []string{"*"},
				// Deny: []string{"order.*"},
			},
		}
	}

	userJWT, err := userClaims.Encode(issuerSeed)
	if err != nil {
		return err
	}
	responseClaims.Jwt = userJWT

	responseToken, err := responseClaims.Encode(issuerSeed)
	if err != nil {
		return err
	}

	encryptedResponse, err := s.encryptedResponse([]byte(responseToken), reqClaims)
	if err != nil {
		return err
	}

	err = msg.Respond(encryptedResponse)
	if err != nil {
		return err
	}
	fmt.Println("Sent authorized response")
	return nil
}

func (s *Server) Handler(msg *nats.Msg) {
	reqClaims, err := s.decryptPayloadMessage(msg)
	if err != nil {
		log.Default().Print("fail to decrypt msg: %w", err)
	}

	for _, user := range s.natsApps {
		if user.Username == reqClaims.AuthorizationRequest.ConnectOptions.Username {
			if user.Password == reqClaims.AuthorizationRequest.ConnectOptions.Password {
				// AUTHORIZED
				// Send authorized response
				if msg.Reply != "" {
					err := s.sendAuthorizedResponse(msg, reqClaims)
					if err != nil {
						log.Default().Print("fail to send authorized response: %w", err)
					}
				}
				return
			}
		}
	}
	// UNAUTHORIZED
	if msg.Reply != "" {
		err := s.sendUnauthorizedResponse(msg, reqClaims)
		if err != nil {
			log.Default().Print("fail to send error response: %w", err)
		}
	}
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
