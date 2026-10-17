package validation

import (
	"context"
	"encoding/json"
	"errors"
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
		if unmarshalTypeErr.Field != "" {
			appErr.Errors = []apperror.ValidationErrorItem{{
				Field:   lowerCamelCase(unmarshalTypeErr.Field),
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
	if err := c.ShouldBindJSON(obj); err != nil {
		return TranslateValidationError(c.Request.Context(), err)
	}
	normalizeNFC(reflect.ValueOf(obj))
	return nil
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
