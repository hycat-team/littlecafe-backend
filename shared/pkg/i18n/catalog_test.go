package i18n

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCatalogValidateBundles(t *testing.T) {
	catalog := MustNewCatalog(DefaultLocale)
	if err := catalog.Validate(); err != nil {
		t.Fatalf("embedded bundles are inconsistent: %v", err)
	}
}

func TestCatalogResolvesLocale(t *testing.T) {
	catalog := MustNewCatalog(DefaultLocale)

	vi := catalog.T(VI, ErrorTitleNotFound)
	en := catalog.T(EN, ErrorTitleNotFound)
	if vi != "Không tìm thấy dữ liệu" {
		t.Fatalf("vi title = %q", vi)
	}
	if en != "Data not found" {
		t.Fatalf("en title = %q", en)
	}
}

func TestCatalogFormatsArgs(t *testing.T) {
	catalog := MustNewCatalog(DefaultLocale)

	if got := catalog.T(VI, ValidationRequired, "email"); got != "Vui lòng nhập email" {
		t.Fatalf("vi formatted = %q", got)
	}
	if got := catalog.T(EN, ValidationRequired, "email"); got != "Please enter email" {
		t.Fatalf("en formatted = %q", got)
	}
}

func TestCatalogFallsBackToDefaultThenKey(t *testing.T) {
	catalog := MustNewCatalog(DefaultLocale)
	catalog.Register(EN, map[string]string{})

	if got := catalog.T(EN, ErrorInternalGeneric); got != "Something went wrong. Please try again later." {
		t.Fatalf("expected en bundle value, got %q", got)
	}
	if got := catalog.T(EN, "missing.key"); got != "missing.key" {
		t.Fatalf("missing key should surface itself, got %q", got)
	}
}

func TestCatalogRegister(t *testing.T) {
	catalog := MustNewCatalog(DefaultLocale)
	catalog.Register(VI, map[string]string{"success.custom": "Thành công"})
	catalog.Register(EN, map[string]string{"success.custom": "Success"})

	if got := catalog.T(VI, "success.custom"); got != "Thành công" {
		t.Fatalf("vi custom = %q", got)
	}
	if got := catalog.T(EN, "success.custom"); got != "Success" {
		t.Fatalf("en custom = %q", got)
	}
}

func TestPackageT(t *testing.T) {
	ctx := WithLocale(context.Background(), EN)
	if got := T(ctx, ErrorAuthRequired); got != "Please sign in" {
		t.Fatalf("T = %q", got)
	}
}

func TestLoadLocaleFSMergesBundles(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/error.json":      {Data: []byte(`{"error.title.not_found":"Data not found"}`)},
		"locales/en/validation.json": {Data: []byte(`{"validation.required":"Please enter %s"}`)},
	}

	messages, err := loadLocaleFS(fsys, EN)
	if err != nil {
		t.Fatalf("loadLocaleFS = %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("merged keys = %d, want 2", len(messages))
	}
	if messages[ErrorTitleNotFound] != "Data not found" {
		t.Fatalf("error bundle value = %q", messages[ErrorTitleNotFound])
	}
	if messages[ValidationRequired] != "Please enter %s" {
		t.Fatalf("validation bundle value = %q", messages[ValidationRequired])
	}
}

func TestLoadLocaleFSRejectsDuplicateKeys(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/first.json":  {Data: []byte(`{"success.ok":"ok"}`)},
		"locales/en/second.json": {Data: []byte(`{"success.ok":"OK"}`)},
	}

	_, err := loadLocaleFS(fsys, EN)
	if err == nil {
		t.Fatal("expected duplicate key error")
	}
	if !strings.Contains(err.Error(), "duplicate message key") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadLocaleFSRejectsMissingLocale(t *testing.T) {
	_, err := loadLocaleFS(fstest.MapFS{}, EN)
	if err == nil {
		t.Fatal("expected missing locale error")
	}
	if !strings.Contains(err.Error(), "no message bundles") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadLocaleFSRejectsInvalidJSON(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/broken.json": {Data: []byte(`{"success.ok":`)},
	}

	_, err := loadLocaleFS(fsys, EN)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode bundle") {
		t.Fatalf("error = %v", err)
	}
}
