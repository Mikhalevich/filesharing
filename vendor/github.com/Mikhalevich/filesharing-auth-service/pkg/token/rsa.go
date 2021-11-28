package token

import (
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	//go:embed cert/public_key.pem
	publicPEMData []byte

	//go:embed cert/private_key.pem
	privatePEMData []byte
)

type RSADecoder struct {
	key *rsa.PublicKey
}

func NewRSADecoder() (*RSADecoder, error) {
	return NewRSADecoderWithPublicKey(publicPEMData)
}

func NewRSADecoderWithPublicKey(pubPEMData []byte) (*RSADecoder, error) {
	block, _ := pem.Decode(pubPEMData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing rsa public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse public key error: %w", err)
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not a public key: %v", key)
	}

	return &RSADecoder{
		key: pub,
	}, nil
}

func (rd *RSADecoder) Decode(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rd.key, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token: unable to parse custom claims")
	}

	return claims, nil
}

type RSAEncoder struct {
	key              *rsa.PrivateKey
	expirationPeriod time.Duration
}

func NewRSAEncoder(ep time.Duration) (*RSAEncoder, error) {
	return NewRSAEncoderWithPrivateKey(privatePEMData, ep)
}

func NewRSAEncoderWithPrivateKey(priPEMData []byte, ep time.Duration) (*RSAEncoder, error) {
	block, _ := pem.Decode(priPEMData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing rsa private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key error: %w", err)
	}

	pri, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not a rsa private key: %v", key)
	}

	return &RSAEncoder{
		key:              pri,
		expirationPeriod: ep,
	}, nil
}

// Encode a user object into a JWT string
func (re *RSAEncoder) Encode(user User) (string, error) {
	expirationTime := time.Now().Add(re.expirationPeriod).Unix()

	claims := CustomClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: expirationTime,
			Issuer:    "filesharing.auth.service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(re.key)
}
