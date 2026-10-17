package i18n

import (
	"context"
	"testing"
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
