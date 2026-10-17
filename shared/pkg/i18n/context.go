package i18n

import "context"

type localeContextKey struct{}

// WithLocale stores the negotiated locale on a standard context. Both Gin
// contexts and gRPC/worker contexts can carry it.
func WithLocale(ctx context.Context, loc Locale) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, localeContextKey{}, Normalize(loc))
}

// FromContext reads the locale stored by WithLocale, falling back to the
// default locale when absent.
func FromContext(ctx context.Context) Locale {
	if ctx == nil {
		return DefaultLocale
	}
	if loc, ok := ctx.Value(localeContextKey{}).(Locale); ok {
		return Normalize(loc)
	}
	return DefaultLocale
}
