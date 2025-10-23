package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type JwtService struct {
	PrivateKey *rsa.PrivateKey // Sign key
	PublicKey  *rsa.PublicKey
}

func NewJwtService(rsaKeyPairFilePath string) (*JwtService, error) {
	// Read rsa-key-pair.pem
	rsaKeyPairBytes, err := os.ReadFile(rsaKeyPairFilePath)
	if err != nil {
		return nil, fmt.Errorf("fail to read rsa key pair file")
	}
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(rsaKeyPairBytes)
	if err != nil {
		return nil, fmt.Errorf("fail to parse rsa private key from pem: %w", err)
	}
	return &JwtService{
		PrivateKey: signKey,
		PublicKey:  &signKey.PublicKey,
	}, nil
}

func (s *JwtService) GenerateJWT(claims *jwt.Claims) (string, error) {
	if claims == nil {
		return "", fmt.Errorf("claim is required")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, *claims)
	// Sign with private key
	signedToken, err := token.SignedString(s.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("fail to sign token with private key: %w", err)
	}
	return signedToken, nil
}

func (s *JwtService) VerifyToken(signedToken string, claims jwt.Claims) (*jwt.Token, error) {
	// Parse the JWT string and store the result in `claims`.
	token, err := jwt.ParseWithClaims(signedToken, claims, func(t *jwt.Token) (interface{}, error) {
		return s.PublicKey, nil
	})
	return token, err
}
