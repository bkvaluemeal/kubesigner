package main

import (
	"context"
	"log"
	"net"
	"time"

	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"

	"encoding/base64"
	"encoding/json"

	"google.golang.org/grpc"

	"math/big"

	apipb "k8s.io/externaljwt/apis/v1alpha1"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	addr string = "@kubesigner"
	tokenTTL int64 = 3600*6 // 6 hours

	// RSA primitives
	//https://en.wikipedia.org/wiki/RSA_cryptosystem#Example
	P int64 = 61
	Q int64 = 53
	N int64 = P * Q
	E int = 17
	D int64 = 413

	// JWT primatives
	typ string = "JWT"
	kid string = "kubesigner"
	alg string = "RS256"
)

var (
	publicKey = rsa.PublicKey{big.NewInt(N), E}
	privateKey = rsa.PrivateKey{
		PublicKey: publicKey,
		D: big.NewInt(D),
		Primes: []*big.Int{big.NewInt(P), big.NewInt(Q)},
	}
)

type server struct {
	apipb.UnimplementedExternalJWTSignerServer
	addr string
}

type jwtHeader struct {
	alg string
	kid string
	typ string
}

func (s *server) Sign(ctx context.Context, req *apipb.SignJWTRequest) (*apipb.SignJWTResponse, error) {
	header, err := json.Marshal(jwtHeader{alg, kid, typ})

	if err != nil {
		return nil, err
	}

	hashed := sha256.Sum256([]byte(req.GetClaims()))
	signature, err := rsa.SignPKCS1v15(nil, &privateKey, crypto.SHA256, hashed[:])

	if err != nil {
		return nil, err
	}

	return &apipb.SignJWTResponse{
		Header: base64.StdEncoding.EncodeToString(header),
		Signature: string(signature),
	}, nil
}

func (s *server) FetchKeys(ctx context.Context, req *apipb.FetchKeysRequest) (*apipb.FetchKeysResponse, error) {
	key, err := x509.MarshalPKIXPublicKey(&publicKey)

	if err != nil {
		return nil, err
	}

	return &apipb.FetchKeysResponse{
		Keys: []*apipb.Key{
			{
				KeyId: kid,
				Key: key,
				ExcludeFromOidcDiscovery: false,
			},
		},
		DataTimestamp: timestamppb.New(time.Now()),
		RefreshHintSeconds: 60,
	}, nil
}

func (s *server) Metadata(ctx context.Context, req *apipb.MetadataRequest) (*apipb.MetadataResponse, error) {
	return &apipb.MetadataResponse{
		MaxTokenExpirationSeconds: tokenTTL,
	}, nil
}

func main() {
	netw := "unix"
	lis, err := net.Listen(netw, addr)

	if err != nil {
		log.Fatalf("net.Listen(%q, %q) failed: %v\n", netw, addr, err)
	}

	s := grpc.NewServer()
	apipb.RegisterExternalJWTSignerServer(s, &server{addr: addr})

	log.Printf("serving on %s\n", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
