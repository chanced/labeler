package labeler

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Stringee is implemented by any value that has a FromString method,
// which parses the “native” format for that value from a string and
// returns a bool value to indicate success (true) or failure (false)
// of parsing.
// Use StringeeStrict if returning an error is preferred.
type Stringee interface {
	FromString(s string) bool
}

// StringeeStrict is implemented by any value that has a FromString method,
// which parses the “native” format for that value from a string.
// The FromString method is used to parse a string, returning an error if
// there was an issue while parsing.
type StringeeStrict interface {
	FromString(s string) error
}

// Labeled is the interface implemented by types with a method GetLabels,
// which returns a map[string]string of labels and values
type Labeled interface {
	GetLabels() map[string]string
}

// LabelerStrict is the interface implemented by types with a SetLabels method, which
// accepts map[string]string and handles assignment of those values, returning true if
// successful or false if there were issues assigning the value.
// An error will be returned from Marshal if v does not contain a tag with `label:"*"`,
// have LabelsField set in Options, or implement either Labeler or Marshaler
type LabelerStrict interface {
	SetLabels(map[string]string) bool
}

// Labeler is the interface implemented by types with a SetLabels method, which
// accepts map[string]string and handles assignment of those values.
// An error will be returned from Marshal if v does not contain a tag with `label:"*"`,
// have LabelsField set in Options, or implement either Labeler or Marshaler
type Labeler interface {
	SetLabels(map[string]string)
}

// Unmarshaler is implemented by any type that has the method UnmarshalLabels,
// providing a means of unmarshaling map[string]string themselves.
type Unmarshaler interface {
	UnmarshalLabels(v map[string]string) error
}

// Marshaler is the interface implemented by types that can marshal a value
// into map[string]string
type Marshaler interface {
	MarshalLabels() error
}

type labelTag struct {
	Key              string
	Default          string
	DefaultWasSet    bool
	IsContainer      bool
	IgnoreCase       bool
	IgnoreCaseWasSet bool
	Required         bool
	RequiredWasSet   bool
	Keep             bool
	KeepWasSet       bool
}

//Unmarshal parses the labels(map[string]string) returned from method GetLabels.
// There must be a means of setting the labels. The top level type must either
// implement Labeler (by having a SetLabels(map[string]string) method), have a
// tag `labels:"*"` on a field at the top level of v, or by having set LabelsField
// in the options (use the Field option func)
func Unmarshal(labeled Labeled, v interface{}, opts ...Option) error {

	o := getOptions(opts)

	labels := map[string]string{}
	// copying labels so the original is not altered with removals
	for lKey, lVal := range labeled.GetLabels() {
		labels[lKey] = lVal
	}

	if m, ok := v.(Unmarshaler); ok {
		return m.UnmarshalLabels(labels)
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidValue
	}

	rv = rv.Elem()
	rvi := rv.Addr().Interface()

	if rv.Kind() != reflect.Struct {
		return ErrInvalidValue
	}
	t := rv.Type()

	errs := []*FieldError{}
	fields := []labelField{}

	hasContainerField := false
	var containerField labelField

	for i := 0; i < rv.NumField(); i++ {
		fieldValue := rv.Field(i)
		structField := t.Field(i)
		fieldType := structField.Type
		fieldName := structField.Name
		fieldKind := fieldValue.Kind()
		iface := fieldValue.Addr().Interface()

		// check to see if the field implements the Unmarshaler interface.
		// If so, skip the tag check and invoke UnmarshalLabels instead.
		if unmarshaler, ok := iface.(Unmarshaler); ok {
			err := unmarshaler.UnmarshalLabels(labels)
			if err != nil {
				errs = append(errs, NewFieldError(fieldName, err))
			}
			continue
		}

		// check to see if the field is a struct. If it is possible to interface
		// without panicing, attempt to traverse by calling Unmarshal with the
		// field

		if fieldKind == reflect.Struct {
			if !fieldValue.Addr().CanInterface() {
				continue
			}
			err := Unmarshal(labeled, iface)
			if err != nil {
				var e *ParsingError
				if errors.Is(err, ErrInvalidValue) {
					// this shouldn't happen. it is ignored for now
					continue
				}
				if errors.As(err, &e) {
					pErrs := []*FieldError{}

					// loop over the errors returned while unmarshaling the struct
					// changing the Field value of each FieldError to {fieldName}.{errField}
					// and adding that to the existing set of errors
					for _, fErr := range e.Errs {
						errField := fmt.Sprintf("%s.%s", fieldName, fErr.Field)
						fieldErr := NewFieldError(errField, fErr)
						errs = append(errs, fieldErr)
					}
					errs = append(errs, pErrs...)
					continue
				}
				// unknown error gets returned for now
				return err
			}

			continue
		}

		var f labelField
		if tagStr, ok := structField.Tag.Lookup("label"); ok {
			if !fieldValue.CanSet() {
				errs = append(errs, NewFieldError(fieldName, ErrUnexportedField))
				continue
			}
			tag, err := parseTag(fieldName, tagStr, *o)
			if err != nil {
				errs = append(errs, NewFieldError(fieldName, ErrMalformedTag))
				continue
			}

			f = labelField{
				Kind:        fieldKind,
				Name:        fieldName,
				Tag:         tag,
				Type:        fieldType,
				Field:       structField,
				Value:       fieldValue,
				Interface:   iface,
				IsContainer: tag.IsContainer,
			}
			fields = append(fields, f)
		} else if o.LabelsField != "" && fieldName == o.LabelsField {
			f = labelField{
				Kind: fieldKind,
				Name: fieldName,
				Tag: labelTag{
					IsContainer:      true,
					DefaultWasSet:    false,
					IgnoreCaseWasSet: false,
					KeepWasSet:       false,
					RequiredWasSet:   false,
				},
				Type:        fieldType,
				Field:       structField,
				Value:       fieldValue,
				Interface:   iface,
				IsContainer: true,
			}
			fields = append(fields, f)
		}
	}

	if len(fields) == 0 {
		if len(errs) > 0 {
			return NewParsingError(errs)
		}
		return nil
	}
	// fields should be cached to improve performance. The loop above and the one immediately following
	// would not need to occur, except in circumstances where there is potentially a field that could not
	// be checked due to being nil.
	// as of right now, all fields are looped over once, those with tags are then looped over again twice.

	// loop over fields to determine if a wildcard has been set, setting the applicable options
	for i, field := range fields {
		tag := field.Tag
		// if there is a container field, denoted by *, then set applicable options
		if field.IsContainer {
			if hasContainerField && containerField.Name != field.Name {
				// if there is already a container field, return an error
				return ErrMultipleContainers
			}

			hasContainerField = true
			if tag.IgnoreCaseWasSet {
				o.IgnoreCase = tag.IgnoreCase
			}
			if tag.RequiredWasSet {
				o.RequireAllFields = tag.Required
			}
			if tag.KeepWasSet {
				o.KeepLabels = tag.Keep
			}
			o.LabelsField = field.Name
			containerField = field
			hasContainerField = true
			fields = append(fields[:i], fields[i+1:]...)
			continue
		}
	}

	for _, field := range fields {
		field.setValue(labels, *o)
	}

	if hasContainerField {
		err := containerField.setValue(labels, *o)
		if err != nil {
			errs = append(errs, NewFieldError(containerField.Name, err))
		}
	} else if iface, ok := rvi.(LabelerStrict); ok {
		if setOk := iface.SetLabels(labels); setOk {
			return ErrSettingLabels
		}
	} else if iface, ok := rvi.(Labeler); ok {
		iface.SetLabels(labels)
	} else {
		return ErrSettingLabels
	}

	if len(errs) > 0 {
		return NewParsingError(errs)
	}
	return nil
}

func getValueFromLabels(m map[string]string, key string, ignoreCase bool) (string, string, bool) {
	if !ignoreCase {
		v, ok := m[key]
		return key, v, ok
	}
	lk := strings.ToLower(key)
	for k, v := range m {
		if lk == strings.ToLower(k) {
			return k, v, true
		}
	}
	return key, "", false
}

func parseTag(fieldName, tagStr string, o Options) (labelTag, error) {
	tag := labelTag{
		IgnoreCaseWasSet: false,
		RequiredWasSet:   false,
		KeepWasSet:       false,
		Required:         false,
		IsContainer:      false,
		IgnoreCase:       o.IgnoreCase,
		Keep:             o.KeepLabels,
	}
	tagStr = strings.TrimSpace(tagStr)
	if tagStr == "" {
		return tag, NewFieldError(fieldName, ErrMalformedTag)
	}

	keys := strings.Split(tagStr, ",")

	for i, key := range keys {
		var k string
		var v string
		if strings.Contains(key, ":") {
			pair := strings.SplitN(key, ":", 2)
			k = strings.ToLower(pair[0])
			v = strings.TrimSpace(pair[1])
		} else {
			k = strings.ToLower(key)
		}
		k = strings.TrimSpace(k)

		switch k {
		case "ignorecase":
			tag.IgnoreCase = true
			tag.IgnoreCaseWasSet = true
		case "casesenstive":
			tag.IgnoreCase = false
			tag.IgnoreCaseWasSet = true
		case "required":
			tag.Required = true
			tag.RequiredWasSet = true
		case "notrequired": // this is overkill
			tag.Required = false
			tag.RequiredWasSet = true
		case "*":
			tag.IsContainer = true
		case "discard":
			tag.Keep = false
			tag.KeepWasSet = false
		case "keep":
			tag.Keep = true
		case "default":
			if v == "" {
				return tag, NewFieldError(fieldName, ErrMalformedTag)
			}
			tag.Default = v
		default:
			tag.Key = strings.TrimSpace(keys[i])
		}

	}
	return tag, nil
}

type labelField struct {
	Tag         labelTag
	Type        reflect.Type
	Kind        reflect.Kind
	Value       reflect.Value
	Field       reflect.StructField
	Interface   interface{}
	Name        string
	IsContainer bool
}

func (field labelField) setValue(labels map[string]string, o Options) error {
	tag := field.Tag
	if field.IsContainer {
		mapType := reflect.TypeOf(map[string]string{})
		if iface, ok := field.Interface.(Unmarshaler); ok {
			err := iface.UnmarshalLabels(labels)
			if err != nil {
				return NewFieldError(field.Name, ErrUnmarshalingLabels)
			}
			return nil
		}
		if iface, ok := field.Interface.(LabelerStrict); ok {
			if setOk := iface.SetLabels(labels); setOk {
				return nil
			}
			return ErrSettingLabels
		}
		if iface, ok := field.Interface.(LabelerStrict); ok {
			if setOk := iface.SetLabels(labels); !setOk {
				return ErrSettingLabels
			}
			return nil
		}
		if iface, ok := field.Interface.(Labeler); ok {
			iface.SetLabels(labels)
			return nil
		}

		if field.Type == mapType && field.Value.CanSet() {
			val := reflect.ValueOf(labels)
			field.Value.Set(val)
			return nil
		}

		return ErrSettingLabels

	}

	var required bool
	var ignoreCase bool
	var keep bool
	var defaultValue string
	if tag.IgnoreCaseWasSet {
		ignoreCase = tag.IgnoreCase
	} else {
		ignoreCase = o.IgnoreCase
	}
	if tag.DefaultWasSet {
		defaultValue = tag.Default
	} else {
		defaultValue = o.Default
	}
	if tag.RequiredWasSet {
		required = tag.Required
	} else {
		required = o.RequireAllFields
	}
	if tag.KeepWasSet {
		keep = tag.Keep
	} else {
		keep = o.KeepLabels
	}

	key, value, ok := getValueFromLabels(labels, tag.Key, ignoreCase)
	if !ok {
		if defaultValue != "" {
			value = defaultValue
		} else if required {
			return NewFieldError(field.Name, ErrMissingRequiredLabel)
		} else {
			return nil
		}
	}

	if iface, isStringee := field.Interface.(Stringee); isStringee {
		if fromStringOk := iface.FromString(value); fromStringOk {
			return handleSet(field.Name, labels, key, keep, nil)
		}
	}
	if iface, ok := field.Interface.(StringeeStrict); ok {
		err := iface.FromString(value)
		return handleSet(field.Name, labels, key, keep, err)
	}
	if iface, ok := field.Interface.(encoding.TextUnmarshaler); ok {
		err := iface.UnmarshalText([]byte(value))
		return handleSet(field.Name, labels, key, keep, err)
	}
	switch field.Kind {
	case reflect.Ptr:
		ptr := reflect.New(field.Type.Elem())
		ptrEl := ptr.Elem()
		ptrField := labelField{
			Value:       ptrEl,
			Kind:        ptrEl.Kind(),
			Type:        field.Type.Elem(),
			Field:       field.Field,
			Interface:   field.Interface,
			IsContainer: field.IsContainer,
			Tag:         field.Tag,
		}
		err := ptrField.setValue(labels, o)
		if err != nil {
			return handleSet(field.Name, labels, key, keep, err)
		}
		field.Value.Set(ptr)
		return handleSet(field.Name, labels, key, keep, err)
	case reflect.String:
		field.Value.SetString(value)
		return handleSet(field.Name, labels, key, keep, nil)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return handleSet(field.Name, labels, key, keep, err)
		}
		field.Value.SetBool(v)
		return handleSet(field.Name, labels, key, keep, nil)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type.PkgPath() == "time" && field.Type.Name() == "Duration" {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return handleSet(field.Name, labels, key, keep, err)
			}
			field.Value.Set(reflect.ValueOf(duration))
			return handleSet(field.Name, labels, key, keep, err)
		}

		v, err := strconv.Atoi(value)
		if err != nil {
			return handleSet(field.Name, labels, key, keep, err)
		}

		field.Value.SetInt(int64(v))

		return handleSet(field.Name, labels, key, keep, err)

	case reflect.Float32, reflect.Float64:
		var v float64
		var err error
		if field.Kind == reflect.Float32 {
			v, err = strconv.ParseFloat(value, 32)
		} else {
			v, err = strconv.ParseFloat(value, 64)
		}
		if err != nil {
			return handleSet(field.Name, labels, key, keep, err)
		}
		field.Value.SetFloat(float64(v))
		return handleSet(field.Name, labels, key, keep, err)
	default:
		return handleSet(field.Name, labels, key, keep, ErrUnsupportedType)
	}
}

func handleSet(name string, labels map[string]string, key string, keep bool, err error) error {
	if !keep && err == nil {
		delete(labels, key)
	}
	return NewFieldError(name, err)
}
