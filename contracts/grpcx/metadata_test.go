package grpcx

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestContextWithAppJWT(t *testing.T) {
	ctx := ContextWithAppJWT(context.Background(), "token")
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("metadata missing")
	}
	if got := md.Get(MetadataAppJWT); len(got) != 1 || got[0] != "token" {
		t.Fatalf("unexpected metadata: %v", got)
	}
}
