package grpcx

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

// MetadataAppJWT carries user identity only when a service acts on behalf of a
// user. mTLS authenticates the calling service at transport level.
const MetadataAppJWT = "x-littlecafe-jwt"

// MetadataLocale carries the negotiated client locale across the gRPC boundary.
const MetadataLocale = "x-littlecafe-locale"

func ContextWithAppJWT(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, MetadataAppJWT, token)
}

// ContextWithLocale stores loc locally and forwards it as outgoing metadata.
func ContextWithLocale(ctx context.Context, loc i18n.Locale) context.Context {
	loc = i18n.Normalize(loc)
	ctx = i18n.WithLocale(ctx, loc)
	return metadata.AppendToOutgoingContext(ctx, MetadataLocale, string(loc))
}

// LocaleFromOutgoingContext reads the locale stored on an outgoing context.
func LocaleFromOutgoingContext(ctx context.Context) i18n.Locale {
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if values := md.Get(MetadataLocale); len(values) > 0 {
			return i18n.Normalize(i18n.Locale(values[0]))
		}
	}
	return i18n.FromContext(ctx)
}

// LocaleFromIncomingContext reads the locale sent by a caller, falling back to
// the default locale when the metadata is absent.
func LocaleFromIncomingContext(ctx context.Context) i18n.Locale {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(MetadataLocale); len(values) > 0 {
			return i18n.Normalize(i18n.Locale(values[0]))
		}
	}
	return i18n.DefaultLocale
}

// ContextWithIncomingLocale binds the incoming metadata locale to ctx so
// downstream code can use i18n.FromContext.
func ContextWithIncomingLocale(ctx context.Context) context.Context {
	return i18n.WithLocale(ctx, LocaleFromIncomingContext(ctx))
}

// LocaleUnaryServerInterceptor propagates the caller's locale into the handler
// context.
func LocaleUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(ContextWithIncomingLocale(ctx), req)
	}
}

// LocaleStreamServerInterceptor propagates the caller's locale into the stream
// handler context.
func LocaleStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &localeServerStream{
			ServerStream: ss,
			ctx:          ContextWithIncomingLocale(ss.Context()),
		})
	}
}

// LocaleUnaryClientInterceptor forwards the locale carried by ctx as outgoing
// metadata.
func LocaleUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(ContextWithLocale(ctx, i18n.FromContext(ctx)), method, req, reply, cc, opts...)
	}
}

// LocaleStreamClientInterceptor forwards the locale carried by ctx as outgoing
// metadata.
func LocaleStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(ContextWithLocale(ctx, i18n.FromContext(ctx)), desc, cc, method, opts...)
	}
}

type localeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *localeServerStream) Context() context.Context {
	return s.ctx
}
