package labeler

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type field struct {
	Tag                    Tag
	Type                   reflect.Type
	Kind                   reflect.Kind
	Value                  reflect.Value
	Field                  reflect.StructField
	Addr                   reflect.Value
	Ptr                    reflect.Value
	Interface              interface{}
	Name                   string
	Key                    string
	TagStr                 string
	WasSet                 bool
	Keep                   bool
	IsTime                 bool
	IsStruct               bool
	IsDuration             bool
	IsPtr                  bool
	IsTagged               bool
	IsContainer            bool
	CanAddr                bool
	CanInterface           bool
	IsSettableMap          bool
	Unmarshaler            Unmarshaler
	UnmarshalerWithOptions UnmarshalerWithOptions
	Marshaler              Marshaler
	Stringer               fmt.Stringer
	Stringee               Stringee
	Labeler                Labeler
	Labeled                Labeled
	GenericallyLabeled     GenericallyLabeled
	StrictLabeler          StrictLabeler
	GenericLabeler         GenericLabeler
	TextUnmarshaler        encoding.TextUnmarshaler
	TextMarshaler          encoding.TextMarshaler
}

var labelMapType reflect.Type = reflect.TypeOf(map[string]string{})

func newField(structField reflect.StructField, fieldValue reflect.Value, o Options) (field, error) {

	fieldName := structField.Name

	isContainer := o.ContainerField != "" && o.ContainerField == fieldName
	canAddr := fieldValue.CanAddr()
	tagStr, isTagged := structField.Tag.Lookup(o.Tag)
	if isTagged {
		isTagged = tagStr != ""
	}
	f := field{
		Name:        fieldName,
		Field:       structField,
		IsContainer: isContainer,
		CanAddr:     canAddr,
		IsTagged:    isTagged,
		TagStr:      tagStr,
	}
	if isTagged && !canAddr {
		return f, f.err(ErrUnexportedField)
	}

	if !canAddr {
		return f, nil
	}

	fieldKind := fieldValue.Kind()

	if fieldKind == reflect.Ptr {
		f.IsPtr = true
		var ptr reflect.Value

		if fieldValue.IsNil() {
			elem := fieldValue.Type().Elem()
			ptr = reflect.New(elem)
		} else {
			ptr = fieldValue.Elem()
		}

		f.Kind = ptr.Elem().Kind()
		f.Value = ptr.Elem()
		f.Ptr = ptr
		fieldValue.Set(ptr)
	} else {
		f.IsPtr = false
		f.Value = fieldValue
		f.Kind = fieldKind
	}

	f.CanInterface = f.Value.CanAddr() && f.Value.Addr().CanInterface()
	if !f.CanInterface {
		return f, nil
	}
	f.Interface = f.Value.Addr().Interface()
	switch t := f.Interface.(type) {
	case UnmarshalerWithOptions:
		f.UnmarshalerWithOptions = t
	case Unmarshaler:
		f.Unmarshaler = t
	case Stringee:
		f.Stringee = t
	case *time.Time:
		f.IsTime = true
	case *time.Duration:
		f.IsDuration = true
	case time.Time:
		f.IsTime = true
	case time.Duration:
		f.IsDuration = true
	case encoding.TextUnmarshaler:
		f.TextUnmarshaler = t
	}

	switch t := f.Interface.(type) {
	case Marshaler:
		f.Marshaler = t
	case GenericallyLabeled:
		f.GenericallyLabeled = t
	case Labeled:
		f.Labeled = t
	case time.Time:
		f.IsTime = true
	case time.Duration:
		f.IsDuration = true
	case *time.Time:
		f.IsTime = true
	case *time.Duration:
		f.IsDuration = true
	case fmt.Stringer:
		f.Stringer = t
	case encoding.TextMarshaler:
		f.TextMarshaler = t
	}

	if f.IsTagged {
		err := f.parseTag(o)
		if err != nil {
			return f, err
		}
	}
	if f.IsContainer {
		switch t := f.Interface.(type) {
		case GenericLabeler:
			f.GenericLabeler = t
		case StrictLabeler:
			f.StrictLabeler = t
		case Labeler:
			f.Labeler = t
		case map[string]string: // may need to make this configurable
			f.IsSettableMap = true
		case *map[string]string: // may need to make this configurable
			f.IsSettableMap = true
		default:
			return f, f.err(ErrInvalidContainer)
		}
		return f, nil
	}
	// not sure if I should check if fieldKind == reflect.Interface
	if !f.IsTime && f.Kind == reflect.Struct {
		f.IsStruct = true
	}

	return f, nil

}

func (f *field) parseTag(o Options) error {
	tagStr := strings.TrimSpace(f.TagStr)
	t, err := NewTag(tagStr, o)
	if err != nil {
		return err
	}
	f.Tag = t
	return nil
}

func (f *field) get(o Options) (map[string]string, error) {

	if !f.IsTagged && !f.IsContainer {
		return make(map[string]string), nil
	}
	l := make(map[string]string)
	key := f.Tag.Key
	switch {
	case f.Marshaler != nil:
		return f.Marshaler.MarshalLabels()
	case f.IsContainer && f.GenericallyLabeled != nil:
		return f.GenericallyLabeled.GetLabels(o.Tag), nil
	case f.IsContainer && f.Labeled != nil:
		return f.Labeled.GetLabels(), nil
	case f.Stringer != nil:
		s := f.Stringer.String()
		l[key] = s
		return l, nil
	case f.TextMarshaler != nil:
		t, err := f.TextMarshaler.MarshalText()
		if err != nil {
			return l, newFieldError(f, err)
		}
		l[key] = string(t)
		return l, nil
	}
	var timeLayout string
	if f.Tag.Format != "" {
		timeLayout = f.Tag.Format
	} else {
		timeLayout = o.TimeFormat
	}

	var err error
	var value string
	switch t := f.Interface.(type) {
	case *time.Time:
		value = t.Format(timeLayout)
	case *time.Duration:
		value = t.String()
	case *string:
		value = *t
	case *bool:
		value = strconv.FormatBool(*t)
	case *int64:
		value = strconv.FormatInt(*t, 64)
	case *int32:
		i64 := int64(*t)
		value = strconv.FormatInt(i64, 32)
	case *int16:
		i64 := int64(*t)
		value = strconv.FormatInt(i64, 16)
	case *int8:
		i64 := int64(*t)
		value = strconv.FormatInt(i64, 8)
	case *uint64:
		value = strconv.FormatUint(*t, 64)
	case *uint32:
		i64 := uint64(*t)
		value = strconv.FormatUint(i64, 32)
	case *uint16:
		i64 := uint64(*t)
		value = strconv.FormatUint(i64, 16)
	case *uint8:
		i64 := uint64(*t)
		value = strconv.FormatUint(i64, 8)
	case *float32:
		f64 := float64(*t)
		var floatFormat byte
		floatFormat, err = f.Tag.GetFloatFormat(o)
		if err == nil {
			value = strconv.FormatFloat(f64, floatFormat, -1, 32)
		}
	case *float64:
		var floatFormat byte
		floatFormat, err = f.Tag.GetFloatFormat(o)
		if err == nil {
			value = strconv.FormatFloat(*t, floatFormat, -1, 32)
		}

	default:
		err = ErrUnsupportedType
	}
	if err != nil {
		return l, f.err(err)
	}
	l[key] = value
	return l, nil
}

func (f *field) set(l map[string]string, o Options) error {
	switch {
	case f.UnmarshalerWithOptions != nil:
		err := f.UnmarshalerWithOptions.UnmarshalLabels(l, o)
		if err != nil {
			return f.err(err)
		}
		return nil
	case f.Unmarshaler != nil:
		err := f.Unmarshaler.UnmarshalLabels(l)
		if err != nil {
			return f.err(err)
		}
		return nil
	case f.IsContainer:
		return f.resolveContainer(l, o)
	case !f.IsTagged:
		return nil
	}
	var required bool
	var ignoreCase bool
	var defaultValue string

	if f.Tag.IgnoreCaseIsSet {
		ignoreCase = f.Tag.IgnoreCase
	} else {
		ignoreCase = o.IgnoreCase
	}

	if f.Tag.DefaultIsSet {
		defaultValue = f.Tag.Default
	} else {
		defaultValue = o.Default
	}

	if f.Tag.RequiredIsSet {
		required = f.Tag.Required
	} else {
		required = o.RequireAllFields
	}

	if f.Tag.KeepIsSet {
		f.Keep = f.Tag.Keep
	} else {
		f.Keep = o.KeepLabels
	}

	key, value, hasKey := f.getKeyAndValueFromLabels(l, ignoreCase)
	if !hasKey {
		if defaultValue != "" {
			value = defaultValue
		} else if required {
			return f.err(ErrMissingRequiredLabel)
		} else {
			return nil
		}
	}
	f.Key = key
	switch {
	case f.Stringee != nil:
		err := f.Stringee.FromString(value)
		return f.resolve(l, nil, err)
	case f.TextUnmarshaler != nil:
		err := f.TextUnmarshaler.UnmarshalText([]byte(value))
		return f.resolve(l, nil, err)
	}

	var valueToSet interface{}
	var err error

	switch f.Interface.(type) {
	case *string:
		valueToSet = value
	case *bool:
		valueToSet, err = strconv.ParseBool(value)
	case *time.Time:
		timeLayout, mErr := f.Tag.GetTimeFormat(o)
		if mErr != nil {
			err = f.err(mErr)
		} else {
			valueToSet, err = time.Parse(timeLayout, value)
		}
	case *time.Duration:
		valueToSet, err = time.ParseDuration(value)
	case *int64:
		valueToSet, err = strconv.ParseInt(value, 10, 64)
	case *int32:
		var v int64
		v, err = strconv.ParseInt(value, 10, 32)
		if err == nil {
			valueToSet = int32(v)
		}
	case *int16:
		var v int64
		v, err = strconv.ParseInt(value, 10, 16)
		if err == nil {
			valueToSet = int16(v)
		}
	case *int8:
		var v int64
		v, err = strconv.ParseInt(value, 10, 8)
		if err == nil {
			valueToSet = int8(v)
		}
	case *int:
		valueToSet, err = strconv.Atoi(value)
	case *float64:
		valueToSet, err = strconv.ParseFloat(value, 64)
	case *float32:
		var v float64
		v, err = strconv.ParseFloat(value, 32)
		if err == nil {
			valueToSet = float32(v)
		}
	case *uint:
		var v uint64
		v, err = strconv.ParseUint(value, 10, 64)
		if err == nil {
			valueToSet = uint(v)
		}
	case *uint64:
		valueToSet, err = strconv.ParseUint(value, 10, 64)
	case *uint32:
		var v uint64
		v, err = strconv.ParseUint(value, 10, 32)
		if err == nil {
			valueToSet = uint32(v)
		}
	case *uint16:
		var v uint64
		v, err = strconv.ParseUint(value, 10, 16)
		if err == nil {
			valueToSet = uint16(v)
		}
	case *uint8:
		var v uint64
		v, err = strconv.ParseUint(value, 10, 8)
		if err == nil {
			valueToSet = uint8(v)
		}
	default:
		err = ErrUnsupportedType
	}
	return f.resolve(l, valueToSet, err)
}

func (f *field) resolveContainer(l map[string]string, o Options) error {
	if !f.IsContainer {
		return nil
	}
	switch {
	case f.GenericLabeler != nil:
		err := f.GenericLabeler.SetLabels(l, o.Tag)
		if err != nil {
			return f.err(err)
		}
		return nil
	case f.StrictLabeler != nil:
		err := f.StrictLabeler.SetLabels(l)
		if err != nil {
			return f.err(err)
		}
		return nil
	case f.Labeler != nil:
		f.Labeler.SetLabels(l)
		return nil
	case f.IsSettableMap:
		f.setValue(l)
		return nil
	default:
		// this is only reached if a field is marked as being a container,
		// either through options or with the wildcard tag and does not have
		// a means of setting the labels. Do not use a container field if the
		// initial value has a SetLabels method.
		return f.err(ErrSettingLabels)
	}
}

func (f *field) err(err error) *FieldError {
	if err != nil {
		return newFieldError(f, err)
	}
	return nil
}

func (f *field) resolve(labels map[string]string, value interface{}, err error) error {
	if err != nil {
		return f.err(err)
	}
	if value != nil {
		f.setValue(value)
		f.WasSet = true
	}
	return nil
}

func (f *field) setValue(v interface{}) {
	rv := reflect.ValueOf(v)
	f.Value.Set(rv)
	if f.IsPtr {
		f.Ptr.Set(f.Value.Elem())
	}
}

func (f *field) getKeyAndValueFromLabels(m map[string]string, ignoreCase bool) (string, string, bool) {
	key := f.Tag.Key
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

func (f field) getRefKind() reflect.Kind {
	return f.Kind
}
func (f field) getRefType() reflect.Type {
	return f.Type
}
func (f field) getRefValue() reflect.Value {
	return f.Value
}
func (f field) isStruct() bool {
	return f.IsStruct
}
func (f field) getRefField() *reflect.StructField {
	// returning ptr so that null check on labeler can be used
	return &f.Field
}

func (f field) getRefNumField() int {
	return f.getRefType().NumField()
}
