package grpcx

import (
	"fmt"
	"net/url"

	"google.golang.org/grpc"
)

// DialService opens an mTLS-authenticated connection. Plaintext is not
// accepted; local development uses certificates from .secrets/certs.
func DialService(rawURL string, files MTLSFiles) (*grpc.ClientConn, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("grpcx: service URL is required")
	}
	target := rawURL
	if parsed, err := url.Parse(rawURL); err == nil && parsed.Host != "" {
		target = parsed.Host
	}
	creds, err := ClientTransportCredentials(files)
	if err != nil {
		return nil, err
	}
	return grpc.NewClient(target,
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(LocaleUnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(LocaleStreamClientInterceptor()),
	)
}
