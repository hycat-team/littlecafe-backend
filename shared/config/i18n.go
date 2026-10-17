package config

import (
	"strings"

	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
)

// I18n holds the localization settings shared by every service.
type I18n struct {
	DefaultLocale    string `mapstructure:"default_locale"`
	SupportedLocales string `mapstructure:"supported_locales"`
}

// Default returns the configured default locale, falling back to the compiled
// default when the configured value is empty or unsupported.
func (c I18n) Default() i18n.Locale {
	loc := i18n.Locale(strings.TrimSpace(c.DefaultLocale))
	if i18n.IsSupported(loc) {
		return loc
	}
	return i18n.DefaultLocale
}

// Supported returns the configured locales, de-duplicated and filtered down to
// locales the message catalog can actually serve.
func (c I18n) Supported() []i18n.Locale {
	raw := strings.TrimSpace(c.SupportedLocales)
	if raw == "" {
		return append([]i18n.Locale(nil), i18n.SupportedLocales...)
	}

	seen := make(map[i18n.Locale]struct{})
	locales := make([]i18n.Locale, 0, len(i18n.SupportedLocales))
	for _, part := range strings.Split(raw, ",") {
		loc := i18n.Locale(strings.TrimSpace(part))
		if !i18n.IsSupported(loc) {
			continue
		}
		if _, exists := seen[loc]; exists {
			continue
		}
		seen[loc] = struct{}{}
		locales = append(locales, loc)
	}
	if len(locales) == 0 {
		return append([]i18n.Locale(nil), i18n.SupportedLocales...)
	}
	return locales
}
