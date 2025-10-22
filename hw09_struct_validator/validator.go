package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var allowedFieldKinds = map[reflect.Kind]struct{}{
	reflect.Int:    {},
	reflect.String: {},
	reflect.Slice:  {},
}

var ErrArgumentNotStructure = errors.New("argument is not a struct")

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for _, e := range v {
		sb.WriteString(fmt.Sprintf("%s: %s\n", e.Field, e.Err.Error()))
	}
	return sb.String()
}

func Validate(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return ErrArgumentNotStructure
	}

	numFields := rv.NumField()
	var validationErrors []ValidationError
	for i := 0; i < numFields; i++ {
		t := rv.Type()
		structField := t.Field(i)

		name := structField.Name
		if !structField.IsExported() {
			validationErrors = append(validationErrors, ValidationError{
				Field: name,
				Err:   fmt.Errorf("field %s is not exported", name),
			})
			continue
		}

		kind := structField.Type.Kind()
		if _, ok := allowedFieldKinds[kind]; !ok {
			validationErrors = append(validationErrors, ValidationError{
				Field: name,
				Err:   fmt.Errorf("unsupported struct kind : %s", kind),
			})
			continue
		}

		if kind == reflect.Slice {
			elemKind := structField.Type.Elem().Kind()
			if elemKind != reflect.Int && elemKind != reflect.String {
				validationErrors = append(validationErrors, ValidationError{
					Field: name,
					Err:   fmt.Errorf("unsupported slice element kind: %s", elemKind),
				})
				continue
			}
		}

		field := rv.Field(i)
		tag := structField.Tag.Get("validate")
		if tag == "" {
			continue
		}
		//nolint:exhaustive
		switch kind {
		case reflect.String:
			validationErrors = append(validationErrors, validateString(field, name, tag)...)
		case reflect.Int:
			validationErrors = append(validationErrors, validateInt(field, name, tag)...)
		case reflect.Slice:
			validationErrors = append(validationErrors, validateSlice(field, name, tag)...)
		default:
		}
	}
	if len(validationErrors) > 0 {
		return ValidationErrors(validationErrors)
	}
	return nil
}

func validateString(f reflect.Value, fieldName string, tag string) []ValidationError {
	rules := strings.Split(tag, "|")

	var errs []ValidationError
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		parts := strings.Split(rule, ":")
		if len(parts) != 2 {
			return append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("invalid validator format: %s", tag),
			})
		}

		validationKey := parts[0]
		value := f.String()
		arg := parts[1]
		switch validationKey {
		case "len":
			n, err := strconv.Atoi(arg)
			if err != nil {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("invalid len value: %w", err),
				})
				continue
			}
			if len(value) != n {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("length must be %d, got %d", n, len(value)),
				})
			}

		case "regexp":
			re, err := regexp.Compile(arg)
			if err != nil {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("invalid regexp: %w", err),
				})
				continue
			}
			if !re.MatchString(value) {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("value does not match regexp %q", arg),
				})
			}

		case "in":
			options := strings.Split(arg, ",")
			found := false
			for _, opt := range options {
				if value == opt {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("value must be one of [%s]", strings.Join(options, ", ")),
				})
			}

		default:
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("unknown validation rule: %s", validationKey),
			})
		}
	}
	return errs
}

func validateInt(f reflect.Value, fieldName string, tag string) []ValidationError {
	rules := strings.Split(tag, "|")

	var errs []ValidationError
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		parts := strings.Split(rule, ":")
		if len(parts) != 2 {
			return append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("invalid validator format: %s", tag),
			})
		}
		validationKey := parts[0]
		value := int(f.Int())
		arg := ""
		if len(parts) == 2 {
			arg = parts[1]
		}
		switch validationKey {
		case "min":
			m, err := strconv.Atoi(arg)
			if err != nil {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("invalid min value: %w", err),
				})
				continue
			}
			if value < m {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("value should me more or equal: %d", m),
				})
			}

		case "max":
			m, err := strconv.Atoi(arg)
			if err != nil {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("invalid max value: %w", err),
				})
				continue
			}
			if value > m {
				errs = append(errs, ValidationError{
					Field: fieldName,
					Err:   fmt.Errorf("value should me less or equal: %d", m),
				})
			}

		case "in":
			options := strings.Split(arg, ",")
			errs = append(errs, checkValInSet(options, value, fieldName)...)

		default:
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("unknown validation rule: %s", validationKey),
			})
		}
	}
	return errs
}

func checkValInSet(options []string, value int, fieldName string) []ValidationError {
	found := false
	var errs []ValidationError
	for _, opt := range options {
		opt = strings.TrimSpace(opt)
		num, err := strconv.Atoi(opt)
		if err != nil {
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("invalid in value %q: %w", opt, err),
			})
			continue
		}
		if value == num {
			found = true
			break
		}
	}
	if !found {
		errs = append(errs, ValidationError{
			fieldName,
			fmt.Errorf("value must be one of [%s], got %d", strings.Join(options, ", "), value),
		})
	}
	return errs
}

func validateSlice(f reflect.Value, fieldName string, tag string) []ValidationError {
	var errs []ValidationError
	elemKind := f.Type().Elem().Kind()
	var vf func(f reflect.Value, fieldName string, tag string) []ValidationError
	if elemKind == reflect.Int {
		vf = validateInt
	} else {
		vf = validateString
	}
	for i := 0; i < f.Len(); i++ {
		elem := f.Index(i)
		subErrs := vf(elem, fieldName, tag)
		for _, e := range subErrs {
			e.Field = fmt.Sprintf("%s[%d]", fieldName, i)
			errs = append(errs, e)
		}
	}
	return errs
}
