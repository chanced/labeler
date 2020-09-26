package labeler

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type labelTag struct {
	Key             string
	Default         string
	DefaultIsSet    bool
	IsContainer     bool
	IgnoreCase      bool
	IgnoreCaseIsSet bool
	Required        bool
	RequiredIsSet   bool
	Keep            bool
	KeepIsSet       bool
	Format          string
}

type fieldRef struct {
	Mutex                  *sync.Mutex
	Tag                    labelTag
	Type                   reflect.Type
	Kind                   reflect.Kind
	Value                  reflect.Value
	Field                  reflect.StructField
	Addr                   reflect.Value
	Ptr                    reflect.Value
	Interface              interface{}
	Name                   string
	TagStr                 string
	IsTime                 bool
	IsStruct               bool
	IsDuration             bool
	IsPtr                  bool
	IsTagged               bool
	IsContainer            bool
	CanAddr                bool
	CanInterface           bool
	IsLabeler              bool
	IsStrictLabeler        bool
	IsSettableMap          bool
	Unmarshaler            Unmarshaler
	UnmarshalerWithOptions UnmarshalerWithOptions
	Marshaler              Marshaler
	Stringer               fmt.Stringer
	Stringee               Stringee
	Labeler                Labeler
	StrictLabeler          StrictLabeler
	GenericLabeler         GenericLabeler
	TextUnmarshaler        encoding.TextUnmarshaler
	TextMarshaler          encoding.TextMarshaler
}

var labelMapType reflect.Type = reflect.TypeOf(map[string]string{})

func newFieldRef(structField reflect.StructField, fieldValue reflect.Value, mutex *sync.Mutex, o *Options) (fieldRef, error) {

	fieldName := structField.Name

	isContainer := o.LabelsField != "" && o.LabelsField == fieldName
	canAddr := fieldValue.CanAddr()
	tagStr, isTagged := structField.Tag.Lookup(o.Tag)

	if isTagged {
		isTagged = tagStr != ""
	}

	f := fieldRef{
		Name:        fieldName,
		Field:       structField,
		IsContainer: isContainer,
		CanAddr:     canAddr,
		IsTagged:    isTagged,
		TagStr:      tagStr,
		Mutex:       mutex,
	}
	if isTagged && !canAddr {
		return f, f.err(ErrUnexportedField)
	}

	if !canAddr {
		return f, nil
	}

	fieldKind := fieldValue.Kind()
	fieldType := structField.Type

	if fieldKind == reflect.Ptr {
		f.IsPtr = true
		var ptr reflect.Value
		elem := fieldType.Elem()
		if fieldValue.IsNil() || fieldType.Kind() != reflect.Struct {
			ptr = reflect.New(elem)
		} else {
			ptr = fieldValue.Elem()
		}
		f.Kind = ptr.Kind()
		f.Value = ptr
		f.Ptr = fieldValue
	} else {
		f.IsPtr = false
		f.Value = fieldValue
		f.Kind = fieldKind
	}

	// not sure if I should check if fieldKind == reflect.Interface

	if !isTagged && f.Kind != reflect.Struct && !f.IsContainer {
		return f, nil
	}

	f.CanInterface = f.Value.Addr().CanInterface()

	f.Interface = fieldValue.Addr().Interface()

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

	if !f.IsTime && f.Kind == reflect.Struct {
		f.IsStruct = true
	}

	return f, nil

}

func (f *fieldRef) parseTag(o *Options) error {
	f.Tag = labelTag{}
	tagStr := strings.TrimSpace(f.TagStr)

	if tagStr == "" || tagStr == o.Seperator {
		return f.err(ErrMalformedTag)
	}

	keys := strings.Split(tagStr, o.Seperator)
	f.Tag.Key = keys[0]

	if f.Tag.Key == o.ContainerToken {
		f.IsContainer = true
	}

	if len(keys) == 1 {
		return nil
	}

	for _, key := range keys[1:] {
		key = strings.TrimSpace(key)
		var k string
		if !o.CaseSensitiveTokens {
			k = strings.ToLower(key)
		} else {
			k = key
		}

		switch k {
		case o.IgnoreCaseToken:
			f.Tag.IgnoreCase = true
			f.Tag.IgnoreCaseIsSet = true
			if f.IsContainer {
				o.IgnoreCase = true
			}
			continue
		case o.CaseSensitiveToken:
			f.Tag.IgnoreCase = false
			f.Tag.IgnoreCaseIsSet = true
			if f.IsContainer {
				o.IgnoreCase = false
			}
			continue
		case o.RequiredToken:
			f.Tag.Required = true
			f.Tag.RequiredIsSet = true
			if f.IsContainer {
				o.RequireAllFields = true
			}
			continue
		case o.NotRequiredToken:
			f.Tag.Required = false
			f.Tag.RequiredIsSet = true
			if f.IsContainer {
				o.RequireAllFields = false
			}
			continue
		case o.DiscardToken:
			f.Tag.Keep = false
			f.Tag.KeepIsSet = true
			if f.IsContainer {
				o.KeepLabels = false
			}
			continue
		case o.KeepToken:
			f.Tag.Keep = true
			f.Tag.KeepIsSet = true
			if f.IsContainer {
				o.KeepLabels = true
			}
			continue
		}
		if strings.Contains(k, o.AssignmentStr) {
			switch {
			case strings.Contains(k, o.LayoutToken):
				sub := strings.SplitN(key, o.AssignmentStr, 2)
				if len(sub) != 2 {
					return f.err(ErrMalformedTag)
				}

				f.Tag.Format = strings.TrimSpace(sub[1])
				fmt.Println(f.Tag.Format)
			case strings.Contains(k, o.DefaultToken):
				sub := strings.SplitN(key, o.AssignmentStr, 2)
				if len(sub) != 2 {
					return f.err(ErrMalformedTag)
				}
				f.Tag.Format = strings.TrimSpace(sub[1])
			}
		}
	}

	return nil
}

func (f *fieldRef) set(l map[string]string, o *Options) error {
	switch {
	case f.UnmarshalerWithOptions != nil:
		err := f.UnmarshalerWithOptions.UnmarshalLabels(l, *o)
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
	}

	if f.IsContainer {
		return f.resolveContainer(l, *o)
	}
	if !f.IsTagged {
		return nil
	}
	var required bool
	var ignoreCase bool
	var keep bool
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
		keep = f.Tag.Keep
	} else {
		keep = o.KeepLabels
	}

	var timeLayout string
	if f.Tag.Format != "" {
		timeLayout = f.Tag.Format
	} else {
		timeLayout = o.TimeLayout
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

	switch {
	case f.Stringee != nil:
		err := f.Stringee.FromString(value)
		return f.resolve(l, key, nil, keep, err)
	case f.TextUnmarshaler != nil:
		err := f.TextUnmarshaler.UnmarshalText([]byte(value))
		return f.resolve(l, key, nil, keep, err)
	}

	var valueToSet interface{}
	var err error

	switch ty := f.Interface.(type) {
	case *string:
		valueToSet = value
	case *bool:
		valueToSet, err = strconv.ParseBool(value)
	case *time.Time:
		if timeLayout == "" {
			return f.err(ErrMissingFormat)
		}
		valueToSet, err = time.Parse(timeLayout, value)
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
	default:
		fmt.Println(ty)
		err = ErrUnsupportedType
	}
	return f.resolve(l, key, valueToSet, keep, err)
}

func (f *fieldRef) resolveContainer(l map[string]string, o Options) error {
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

func (f *fieldRef) err(err error) *FieldError {
	if err != nil {
		return newFieldError(f, err)
	}
	return nil
}

func (f *fieldRef) resolve(labels map[string]string, key string, value interface{}, keep bool, err error) error {
	if err != nil {
		return f.err(err)
	}
	if value != nil {
		f.setValue(value)
	}
	if !keep {
		f.Mutex.Lock()
		delete(labels, key)
		f.Mutex.Unlock()
	}
	return nil
}

func (f *fieldRef) setValue(v interface{}) {
	rv := reflect.ValueOf(v)
	f.Value.Set(rv)
	if f.IsPtr {
		f.Ptr.Set(f.Value)
	}
}

func (f *fieldRef) getKeyAndValueFromLabels(m map[string]string, ignoreCase bool) (string, string, bool) {
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
