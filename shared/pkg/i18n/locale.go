package i18n

import (
	"strings"

	"golang.org/x/text/language"
)

// Locale is a supported client-facing language code.
type Locale string

const (
	VI Locale = "vi"
	EN Locale = "en"
)

// DefaultLocale is used when no Accept-Language header is present or the
// requested language is not supported.
const DefaultLocale = VI

// SupportedLocales lists every locale the shared message catalog ships with.
// Keep this in sync with the files under locales/.
var SupportedLocales = []Locale{VI, EN}

var supportedTags = []language.Tag{language.Vietnamese, language.English}

// IsSupported reports whether loc is a locale the catalog can serve.
func IsSupported(loc Locale) bool {
	for _, supported := range SupportedLocales {
		if supported == loc {
			return true
		}
	}
	return false
}

// Normalize falls back to the default locale when loc is unsupported.
func Normalize(loc Locale) Locale {
	if IsSupported(loc) {
		return loc
	}
	return DefaultLocale
}

// Negotiator maps Accept-Language headers and stored preferences to a supported
// locale. It is safe for concurrent use.
type Negotiator struct {
	defaultLocale Locale
	supported     []Locale
	matcher       language.Matcher
}

// NewNegotiator builds a negotiator. Unsupported entries are dropped; when no
// supported locale remains the code defaults are used.
func NewNegotiator(defaultLocale Locale, supported []Locale) *Negotiator {
	locales := make([]Locale, 0, len(supported))
	tags := make([]language.Tag, 0, len(supported))
	for _, loc := range supported {
		tag, ok := tagFor(loc)
		if !ok {
			continue
		}
		locales = append(locales, loc)
		tags = append(tags, tag)
	}
	if len(locales) == 0 {
		locales = append(locales, SupportedLocales...)
		for _, loc := range SupportedLocales {
			tag, _ := tagFor(loc)
			tags = append(tags, tag)
		}
	}
	if !containsLocale(locales, defaultLocale) {
		defaultLocale = locales[0]
	}
	return &Negotiator{
		defaultLocale: defaultLocale,
		supported:     locales,
		matcher:       language.NewMatcher(tags),
	}
}

// DefaultNegotiator negotiates against the compiled-in defaults.
func DefaultNegotiator() *Negotiator {
	return NewNegotiator(DefaultLocale, SupportedLocales)
}

// Parse maps a single language tag (for example a stored user preference) to a
// supported locale.
func (n *Negotiator) Parse(tag string) (Locale, bool) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return n.defaultLocale, false
	}
	parsed, err := language.Parse(tag)
	if err != nil {
		return n.defaultLocale, false
	}
	return n.match([]language.Tag{parsed})
}

// Negotiate resolves an Accept-Language header to a supported locale.
func (n *Negotiator) Negotiate(header string) (Locale, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return n.defaultLocale, false
	}
	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil || len(tags) == 0 {
		return n.defaultLocale, false
	}
	return n.match(tags)
}

// DefaultLocale returns the fallback locale of this negotiator.
func (n *Negotiator) DefaultLocale() Locale {
	if n == nil {
		return DefaultLocale
	}
	return n.defaultLocale
}

func (n *Negotiator) match(tags []language.Tag) (Locale, bool) {
	if n == nil || len(tags) == 0 {
		return DefaultLocale, false
	}
	_, idx, confidence := n.matcher.Match(tags...)
	if confidence == language.No || idx < 0 || idx >= len(n.supported) {
		return n.defaultLocale, false
	}
	return n.supported[idx], true
}

func tagFor(loc Locale) (language.Tag, bool) {
	switch loc {
	case VI:
		return language.Vietnamese, true
	case EN:
		return language.English, true
	default:
		return language.Und, false
	}
}

func containsLocale(locales []Locale, target Locale) bool {
	for _, loc := range locales {
		if loc == target {
			return true
		}
	}
	return false
}
