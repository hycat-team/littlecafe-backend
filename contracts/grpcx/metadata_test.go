package grpcx

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
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

func TestContextWithLocaleRoundTrip(t *testing.T) {
	ctx := ContextWithLocale(context.Background(), i18n.EN)

	if got := i18n.FromContext(ctx); got != i18n.EN {
		t.Fatalf("FromContext = %s, want en", got)
	}
	if got := LocaleFromOutgoingContext(ctx); got != i18n.EN {
		t.Fatalf("LocaleFromOutgoingContext = %s, want en", got)
	}
}

func TestLocaleFromIncomingContext(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataLocale, "en"))

	if got := LocaleFromIncomingContext(ctx); got != i18n.EN {
		t.Fatalf("LocaleFromIncomingContext = %s, want en", got)
	}
	if got := i18n.FromContext(ContextWithIncomingLocale(ctx)); got != i18n.EN {
		t.Fatalf("bound context locale = %s, want en", got)
	}
	if got := LocaleFromIncomingContext(context.Background()); got != i18n.DefaultLocale {
		t.Fatalf("missing metadata = %s, want default", got)
	}
}

func TestLocaleUnaryServerInterceptor(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataLocale, "en"))
	called := false

	_, err := LocaleUnaryServerInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		called = true
		if got := i18n.FromContext(ctx); got != i18n.EN {
			t.Fatalf("handler locale = %s, want en", got)
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}
