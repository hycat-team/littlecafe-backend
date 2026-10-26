package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"
)

//go:embed locales
var localeFS embed.FS

// Catalog holds every client-facing message per locale. Bundles are embedded at
// build time so a missing or partial deployment cannot silently drop strings.
type Catalog struct {
	mu            sync.RWMutex
	defaultLocale Locale
	messages      map[Locale]map[string]string
}

// NewCatalog loads the embedded bundles for the supported locales.
func NewCatalog(defaultLocale Locale) (*Catalog, error) {
	c := &Catalog{
		defaultLocale: Normalize(defaultLocale),
		messages:      make(map[Locale]map[string]string, len(SupportedLocales)),
	}
	for _, loc := range SupportedLocales {
		msgs, err := loadLocale(loc)
		if err != nil {
			return nil, err
		}
		c.messages[loc] = msgs
	}
	return c, nil
}

// MustNewCatalog is NewCatalog that panics on failure; use it during bootstrap
// where a broken bundle must stop the service immediately.
func MustNewCatalog(defaultLocale Locale) *Catalog {
	c, err := NewCatalog(defaultLocale)
	if err != nil {
		panic(fmt.Sprintf("i18n: %v", err))
	}
	return c
}

var (
	defaultOnce    sync.Once
	defaultCatalog *Catalog
	defaultErr     error
)

// Default returns the process-wide catalog built from the embedded bundles.
func Default() *Catalog {
	defaultOnce.Do(func() {
		defaultCatalog, defaultErr = NewCatalog(DefaultLocale)
		if defaultErr == nil {
			defaultErr = defaultCatalog.Validate()
		}
	})
	if defaultErr != nil {
		panic(fmt.Sprintf("i18n: %v", defaultErr))
	}
	return defaultCatalog
}

// T resolves key for the locale carried by ctx.
func T(ctx context.Context, key string, args ...any) string {
	return Default().T(FromContext(ctx), key, args...)
}

// T resolves key for loc, falling back to the default locale and finally the
// key itself so a missing translation is observable instead of blank.
func (c *Catalog) T(loc Locale, key string, args ...any) string {
	template, ok := c.lookup(loc, key)
	if !ok {
		return key
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

// Has reports whether loc (or the default locale) defines key.
func (c *Catalog) Has(loc Locale, key string) bool {
	_, ok := c.lookup(loc, key)
	return ok
}

// Register merges extra messages into a locale, allowing a service to own its
// business messages while reusing the shared catalog.
func (c *Catalog) Register(loc Locale, messages map[string]string) {
	loc = Normalize(loc)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.messages[loc] == nil {
		c.messages[loc] = make(map[string]string, len(messages))
	}
	for key, value := range messages {
		c.messages[loc][key] = value
	}
}

// Validate reports keys that are missing from any supported locale, and key
// sets that drifted between locales.
func (c *Catalog) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.messages) == 0 {
		return fmt.Errorf("no message bundles loaded")
	}

	union := make(map[string]struct{})
	for _, bundle := range c.messages {
		for key := range bundle {
			union[key] = struct{}{}
		}
	}

	var problems []string
	for key := range union {
		for _, loc := range SupportedLocales {
			if _, ok := c.messages[loc][key]; !ok {
				problems = append(problems, fmt.Sprintf("%s missing in %s", key, loc))
			}
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("i18n bundle mismatch: %s", strings.Join(problems, "; "))
}

func (c *Catalog) lookup(loc Locale, key string) (string, bool) {
	loc = Normalize(loc)
	c.mu.RLock()
	defer c.mu.RUnlock()

	if bundle := c.messages[loc]; bundle != nil {
		if value, ok := bundle[key]; ok {
			return value, true
		}
	}
	if bundle := c.messages[c.defaultLocale]; bundle != nil {
		if value, ok := bundle[key]; ok {
			return value, true
		}
	}
	return "", false
}

func loadLocale(loc Locale) (map[string]string, error) {
	return loadLocaleFS(localeFS, loc)
}

// loadLocaleFS merges every bundle in locales/<locale>/ and rejects duplicate
// keys, so a key can only be owned by one namespace file.
func loadLocaleFS(fsys fs.FS, loc Locale) (map[string]string, error) {
	loc = Normalize(loc)
	pattern := path.Join("locales", string(loc), "*.json")
	paths, err := fs.Glob(fsys, pattern)
	if err != nil {
		return nil, fmt.Errorf("glob bundles %s: %w", pattern, err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no message bundles for locale %s", loc)
	}
	sort.Strings(paths)

	messages := make(map[string]string)
	sources := make(map[string]string)
	for _, file := range paths {
		bundle, err := loadBundleFS(fsys, file)
		if err != nil {
			return nil, err
		}
		for key, value := range bundle {
			if previous, ok := sources[key]; ok {
				return nil, fmt.Errorf("duplicate message key %q in %s and %s", key, previous, file)
			}
			sources[key] = file
			messages[key] = value
		}
	}
	return messages, nil
}

func loadBundleFS(fsys fs.FS, name string) (map[string]string, error) {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read bundle %s: %w", name, err)
	}
	messages := make(map[string]string)
	if err := json.Unmarshal(raw, &messages); err != nil {
		return nil, fmt.Errorf("decode bundle %s: %w", name, err)
	}
	return messages, nil
}
