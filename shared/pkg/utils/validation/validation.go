package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/hycat-team/littlecafe-backend/shared/application/constants/apperror"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/i18n"
	"github.com/hycat-team/littlecafe-backend/shared/pkg/utils/numberutils"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", IsUserName)
		v.RegisterValidation("neqfield", NotEqualField)
		v.RegisterValidation("password_complexity", IsPasswordComplex)
		v.RegisterValidation("nfcmax", NFCMax)
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "" {
				name = lowerCamelCase(fld.Name)
			}
			if name == "-" {
				return ""
			}
			label := fld.Tag.Get("label")
			if label != "" {
				return name + "|" + label
			}
			return name
		})
	}
}

func humanizeFieldName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	name = strings.ReplaceAll(name, "_", " ")
	runes := []rune(name)
	var builder strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			if unicode.IsLower(prev) || unicode.IsDigit(prev) {
				builder.WriteRune(' ')
			}
		}
		builder.WriteRune(unicode.ToLower(r))
	}

	return strings.TrimSpace(builder.String())
}

func fieldJSONName(fe validator.FieldError) string {
	// Namespace uses the tag names registered in init (client-facing JSON
	// names) and keeps the parent/index path, e.g. "request.items[0].quantity".
	// Strip the root struct name so clients receive a path relative to the body.
	if namespace := strings.TrimSpace(fe.Namespace()); namespace != "" {
		if idx := strings.Index(namespace, "."); idx != -1 {
			namespace = namespace[idx+1:]
		}
		segments := strings.Split(namespace, ".")
		for i, segment := range segments {
			if idx := strings.Index(segment, "|"); idx != -1 {
				segments[i] = segment[:idx]
			}
		}
		if path := strings.Join(segments, "."); path != "" {
			return path
		}
	}

	field := strings.TrimSpace(fe.Field())
	if field != "" {
		if idx := strings.Index(field, "|"); idx != -1 {
			return field[:idx]
		}
		return field
	}
	return lowerCamelCase(fe.StructField())
}

func fieldLabel(ctx context.Context, fe validator.FieldError) string {
	field := strings.TrimSpace(fe.Field())
	if field != "" {
		if idx := strings.Index(field, "|"); idx != -1 {
			return field[idx+1:]
		}
		return localizeFieldName(ctx, field)
	}
	return localizeFieldName(ctx, fe.StructField())
}

func localizeFieldName(ctx context.Context, name string) string {
	jsonName := lowerCamelCase(name)
	key := i18n.FieldLabelKey(jsonName)
	if i18n.Default().Has(i18n.FromContext(ctx), key) {
		return i18n.T(ctx, key)
	}
	return humanizeFieldName(name)
}

func lowerCamelCase(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			if parts[i] == "" {
				continue
			}
			if i == 0 {
				parts[i] = strings.ToLower(parts[i][:1]) + parts[i][1:]
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, "")
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func validationMessage(ctx context.Context, fe validator.FieldError, fieldName, paramName string) string {
	switch fe.Tag() {
	case "required":
		return i18n.T(ctx, i18n.ValidationRequired, fieldName)
	case "email":
		return i18n.T(ctx, i18n.ValidationEmail)
	case "username":
		return i18n.T(ctx, i18n.ValidationUsername)
	case "password_complexity":
		return i18n.T(ctx, i18n.ValidationPasswordComplexity, fieldName)
	case "eqfield":
		return i18n.T(ctx, i18n.ValidationEqField, fieldName, paramName)
	case "nefield":
		return i18n.T(ctx, i18n.ValidationNeField, fieldName, paramName)
	case "len":
		return i18n.T(ctx, i18n.ValidationLen, fieldName, paramName)
	case "gt":
		return i18n.T(ctx, i18n.ValidationGT, fieldName, paramName)
	case "min":
		return i18n.T(ctx, i18n.ValidationMin, fieldName, paramName)
	case "max":
		if fe.Kind().String() == "int32" || fe.Kind().String() == "int" {
			return i18n.T(ctx, i18n.ValidationMaxInt, fieldName, paramName)
		}
		return i18n.T(ctx, i18n.ValidationMaxStr, fieldName, paramName)
	case "nfcmax":
		return i18n.T(ctx, i18n.ValidationMaxStr, fieldName, paramName)
	case "datetime":
		return i18n.T(ctx, i18n.ValidationDateTime, fieldName, paramName)
	case "oneof":
		return i18n.T(ctx, i18n.ValidationOneOf, fieldName, paramName)
	case "numeric":
		return i18n.T(ctx, i18n.ValidationNumeric, fieldName)
	default:
		return i18n.T(ctx, i18n.ValidationDefault, fieldName, fe.Tag())
	}
}

// TranslateValidationError maps validator and binding errors to the shared
// problem shape. Messages are rendered for the locale carried by ctx.
func TranslateValidationError(ctx context.Context, err error) *apperror.Error {
	return translateValidationError(ctx, err, nil)
}

func translateValidationError(ctx context.Context, err error, raw []byte) *apperror.Error {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		items := make([]apperror.ValidationErrorItem, 0, len(ve))
		for _, fe := range ve {
			fieldName := fieldLabel(ctx, fe)
			paramName := fe.Param()
			msg := validationMessage(ctx, fe, fieldName, paramName)
			items = append(items, apperror.ValidationErrorItem{
				Field:   fieldJSONName(fe),
				Message: msg,
			})
		}

		appErr := apperror.NewBadRequest("")
		if len(items) > 0 {
			appErr.Message = items[0].Message
			appErr.Detail = items[0].Message
			appErr.Errors = items
		}
		return appErr
	}

	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		appErr := apperror.NewError(
			http.StatusBadRequest,
			i18n.ErrorValidationInvalidFormatTitle,
			i18n.ErrorValidationInvalidFormatDetail,
		)
		field := lowerCamelCase(unmarshalTypeErr.Field)
		if len(raw) > 0 {
			if path, ok := jsonPathAtOffset(raw, unmarshalTypeErr.Offset); ok {
				field = path
			}
		}
		if field != "" {
			appErr.Errors = []apperror.ValidationErrorItem{{
				Field:   field,
				Message: appErr.Message,
			}}
		}
		return appErr
	}

	if strings.Contains(err.Error(), "unmarshal") {
		return apperror.NewError(
			http.StatusBadRequest,
			i18n.ErrorValidationInvalidFormatTitle,
			i18n.ErrorValidationInvalidFormatDetail,
		)
	}

	// Bodies exceeding the tiered limit are capped by http.MaxBytesReader
	// (spec 019 FR-002); the read failure surfaces here as *http.MaxBytesError.
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return apperror.NewRequestEntityTooLarge(
			i18n.ValidationBodyTooLarge,
			numberutils.FormatBytes(maxBytesErr.Limit),
		)
	}

	return apperror.NewBadRequest(err.Error())
}

func BindJSON(c *gin.Context, obj any) error {
	raw, err := c.GetRawData()
	if err != nil {
		return TranslateValidationError(c.Request.Context(), err)
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	if err := c.ShouldBindJSON(obj); err != nil {
		return translateValidationError(c.Request.Context(), err, raw)
	}
	normalizeNFC(reflect.ValueOf(obj))
	return nil
}

// jsonPathAtOffset resolves the JSON path of the value ending at target. The
// standard json.UnmarshalTypeError only reports "items.quantity" for arrays, so
// the raw body is scanned to recover array indexes for errors[].field.
func jsonPathAtOffset(raw []byte, target int64) (string, bool) {
	if target <= 0 {
		return "", false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	return scanJSONPath(dec, target, "")
}

func scanJSONPath(dec *json.Decoder, target int64, path string) (string, bool) {
	token, err := dec.Token()
	if err != nil {
		return "", false
	}

	if delim, ok := token.(json.Delim); ok {
		switch delim {
		case '{':
			for dec.More() {
				keyToken, err := dec.Token()
				if err != nil {
					return "", false
				}
				key, _ := keyToken.(string)
				childPath := key
				if path != "" {
					childPath = path + "." + key
				}
				if found, ok := scanJSONPath(dec, target, childPath); ok {
					return found, true
				}
			}
			_, _ = dec.Token()
		case '[':
			for index := 0; dec.More(); index++ {
				childPath := fmt.Sprintf("%s[%d]", path, index)
				if found, ok := scanJSONPath(dec, target, childPath); ok {
					return found, true
				}
			}
			_, _ = dec.Token()
		}
		return "", false
	}

	if int64(dec.InputOffset()) >= target {
		return path, true
	}
	return "", false
}

func BindQuery(c *gin.Context, obj any) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return TranslateValidationError(c.Request.Context(), err)
	}
	return nil
}

func BindURI(c *gin.Context, obj any) error {
	if err := c.ShouldBindUri(obj); err != nil {
		return TranslateValidationError(c.Request.Context(), err)
	}
	return nil
}

func GetUUIDParam(c *gin.Context, name string) (uuid.UUID, error) {
	raw := c.Param(name)
	if raw == "" {
		return uuid.Nil, apperror.NewBadRequest(i18n.ValidationParamMissing, name)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.NewBadRequest(i18n.ValidationParamInvalid, name)
	}
	return id, nil
}

func GetStringParam(c *gin.Context, name string) (string, error) {
	raw := c.Param(name)
	if raw == "" {
		return "", apperror.NewBadRequest(i18n.ValidationParamMissing, name)
	}
	return raw, nil
}

func GetInt64Param(c *gin.Context, name string) (int64, error) {
	raw := c.Param(name)
	if raw == "" {
		return 0, apperror.NewBadRequest(i18n.ValidationParamMissing, name)
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, apperror.NewBadRequest(i18n.ValidationParamInvalid, name)
	}
	return val, nil
}
