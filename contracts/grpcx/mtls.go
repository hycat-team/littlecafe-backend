package grpcx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type MTLSFiles struct{ CAFile, CertFile, KeyFile, ServerName string }

func loadPool(path string) (*x509.CertPool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(raw) {
		return nil, fmt.Errorf("grpcx: invalid CA PEM")
	}
	return pool, nil
}

func ClientTransportCredentials(files MTLSFiles) (credentials.TransportCredentials, error) {
	pool, err := loadPool(files.CAFile)
	if err != nil {
		return nil, fmt.Errorf("grpcx: load client CA: %w", err)
	}
	cert, err := tls.LoadX509KeyPair(files.CertFile, files.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("grpcx: load client certificate: %w", err)
	}
	return credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13, RootCAs: pool, Certificates: []tls.Certificate{cert}, ServerName: files.ServerName}), nil
}

func ServerOption(files MTLSFiles) (grpc.ServerOption, error) {
	pool, err := loadPool(files.CAFile)
	if err != nil {
		return nil, fmt.Errorf("grpcx: load server CA: %w", err)
	}
	cert, err := tls.LoadX509KeyPair(files.CertFile, files.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("grpcx: load server certificate: %w", err)
	}
	creds := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool, Certificates: []tls.Certificate{cert}})
	return grpc.Creds(creds), nil
}

func ServerOptionsFromEnv() ([]grpc.ServerOption, error) {
	options := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(LocaleUnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(LocaleStreamServerInterceptor()),
	}

	files := MTLSFiles{CAFile: os.Getenv("GRPC_TLS_CA_FILE"), CertFile: os.Getenv("GRPC_TLS_CERT_FILE"), KeyFile: os.Getenv("GRPC_TLS_KEY_FILE")}
	if files.CAFile == "" && files.CertFile == "" && files.KeyFile == "" {
		if os.Getenv("GRPC_MTLS_REQUIRED") == "true" {
			return nil, fmt.Errorf("grpcx: mTLS is required but certificate paths are missing")
		}
		return options, nil
	}
	if files.CAFile == "" || files.CertFile == "" || files.KeyFile == "" {
		return nil, fmt.Errorf("grpcx: incomplete mTLS certificate configuration")
	}
	option, err := ServerOption(files)
	if err != nil {
		return nil, err
	}
	return append(options, option), nil
}
