package grpcx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// mTLS authenticates the calling service at transport level. MetadataAppJWT
// carries user identity only when a service acts on behalf of a user.
const MetadataAppJWT = "x-littlecafe-jwt"

func ContextWithAppJWT(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, MetadataAppJWT, token)
}
