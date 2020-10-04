package labeler

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type field struct {
	meta
	Tag         *Tag
	Parent      reflected
	Name        string
	Path        string
	Key         string
	WasSet      bool
	Keep        bool
	IsTagged    bool
	IsContainer bool
}

func newField(sf reflect.StructField, rv reflect.Value, parent reflected, o Options) (*field, error) {
	fieldName := sf.Name
	f := &field{
		Name:   fieldName,
		meta:   newMeta(rv),
		Parent: parent,
	}
	f.parseTag(sf, o)
	f.deref()
	f.setIsContainer(o)

	if f.IsTagged && !f.CanAddr {
		return f, f.err(ErrUnexportedField)
	}
	if !f.CanAddr {
		return f, nil
	}
	return f, nil
}

func (f *field) ignoreCase(o Options) bool {
	if f.Tag.IgnoreCaseIsSet {
		return f.Tag.IgnoreCase
	}
	return o.IgnoreCase
}

func (f *field) parseTag(sf reflect.StructField, o Options) error {
	tagstr, isTagged := sf.Tag.Lookup(o.Tag)

	if !isTagged {
		f.IsTagged = false
		return nil
	}
	t, err := newTag(tagstr, o)
	f.Tag = t
	if err != nil {
		return err
	}
	return nil
}

func (f *field) setIsContainer(o Options) {
	switch {
	case o.ContainerToken == f.Tag.Key:
		f.IsContainer = true
	case o.ContainerField != "" && o.ContainerField == f.Path:
		f.IsContainer = true
	}
	f.IsContainer = false
}

func (f *field) setPath() {
	f.Path = fmt.Sprintf("%s.%s", f.Parent.path(), f.Name)
}

func (f *field) path() string {
	return f.Path
}

func (f *field) err(err error) *FieldError {
	if err != nil {
		return newFieldError(f, err)
	}
	return nil
}

func (f *field) resolve(labels map[string]string, v interface{}, err error) error {
	if err != nil {
		return f.err(err)
	}
	if v != nil {
		f.setValue(v)
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

func (f *field) intBase(o Options) int {
	if base, ok := f.Tag.GetIntBase(); ok {
		return base
	}
	return o.IntBase
}
func (f *field) uintBase(o Options) int {
	if base, ok := f.Tag.GetUintBase(); ok {
		return base
	}
	return o.UintBase
}

func (f *field) floatFormat(o Options) byte {
	if format, ok := f.Tag.GetFloatFormat(); ok {
		return format
	}
	return o.FloatFormat
}

func (f *field) complexFormat(o Options) byte {
	if format, ok := f.Tag.GetComplexFormat(); ok {
		return format
	}
	return o.ComplexFormat
}

func (f *field) timeFormat(o Options) string {
	if format, ok := f.Tag.GetTimeFormat(); ok {
		return format
	}
	return o.TimeFormat
}

func (f *field) formatInt(o Options) (string, bool) {
	switch f.Kind {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		v := f.Value.Int()
		return strconv.FormatInt(v, f.intBase(o)), true
	default:
		return "", false
	}
}

func (f *field) setInt(s string, bits int, o Options) error {
	v, err := strconv.ParseInt(s, f.intBase(o), bits)
	if err != nil {
		return f.err(err)
	}
	f.Value.SetInt(v)
	return nil
}

func (f *field) formatString(o Options) (string, bool) {
	return f.Value.String(), true
}

func (f *field) setString(s string, o Options) error {
	f.Value.SetString(s)
	return nil
}

func (f *field) formatUint(o Options) (string, bool) {
	switch f.Kind {
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		v := f.Value.Uint()
		return strconv.FormatUint(v, f.intBase(o)), true
	default:
		return "", false
	}
}

func (f *field) setUint(s string, bits int, o Options) error {
	v, err := strconv.ParseUint(s, f.uintBase(o), bits)
	if err != nil {
		return f.err(err)
	}
	f.Value.SetUint(v)
	return nil
}

func (f *field) formatComplex(o Options) (string, bool) {
	switch f.Kind {
	case reflect.Complex128:
		v := f.Value.Complex()
		return strconv.FormatComplex(v, f.complexFormat(o), -1, 128), true
	case reflect.Complex64:
		v := f.Value.Complex()
		return strconv.FormatComplex(v, f.complexFormat(o), -1, 64), true
	default:
		return "", false
	}
}

func (f *field) setComplex(s string, bits int, o Options) error {
	v, err := strconv.ParseComplex(s, bits)
	if err != nil {
		return f.err(err)
	}
	f.Value.SetComplex(v)
	return nil
}

func (f *field) formatFloat(o Options) (string, bool) {
	switch f.Kind {
	case reflect.Float64:
		v := f.Value.Float()
		return strconv.FormatFloat(v, f.floatFormat(o), -1, 64), true
	case reflect.Float32:
		v := f.Value.Float()
		return strconv.FormatFloat(v, f.floatFormat(o), -1, 32), true
	default:
		return "", false
	}
}

func (f *field) setFloat(s string, bits int, o Options) error {
	v, err := strconv.ParseFloat(s, bits)
	if err != nil {
		return f.err(err)
	}
	f.Value.SetFloat(v)
	return nil
}

func (f *field) formatBool(o Options) (string, bool) {
	if f.Kind == reflect.Bool {
		v := f.Value.Bool()
		return strconv.FormatBool(v), true
	}
	return "", false
}

func (f *field) setBool(s string, o Options) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return f.err(err)
	}
	f.Value.SetBool(v)
	return nil
}

func (f *field) formatTime(o Options) (string, bool) {
	if v, ok := f.Value.Interface().(time.Time); ok {
		return v.Format(f.timeFormat(o)), true
	}
	return "", false
}

func (f *field) setTime(s string, o Options) error {
	if !timeType.AssignableTo(f.Type) {
		return f.err(errors.New("Can not assign time.Time to " + f.Name))
	}
	v, err := time.Parse(f.timeFormat(o), s)
	if err != nil {
		return f.err(err)
	}
	rv := reflect.ValueOf(v)
	f.Value.Set(rv)
	return nil
}

func (f *field) formatDuration(o Options) (string, bool) {
	if v, ok := f.Value.Interface().(time.Duration); ok {
		return v.String(), true
	}
	return "", false

}

func (f *field) setDuration(s string, o Options) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return f.err(err)
	}
	rv := reflect.ValueOf(v)
	f.Value.Set(rv)
	return nil
}

func (f *field) topic() topic {
	return fieldTopic
}

func (f *field) Save() {
	f.save()
	f.Parent.Save()
}

var labelMapType reflect.Type = reflect.TypeOf(map[string]string{})
var timeType = reflect.TypeOf(time.Time{})
var durationType = func() reflect.Type { var d time.Duration; return reflect.TypeOf(d) }()

// switch {
// case f.UnmarshalerWithOpts != nil:
// 	err := f.UnmarshalerWithOpts.UnmarshalLabels(l, o)
// 	if err != nil {
// 		return f.err(err)
// 	}
// 	return nil
// case f.Unmarshaler != nil:
// 	err := f.Unmarshaler.UnmarshalLabels(l)
// 	if err != nil {
// 		return f.err(err)
// 	}
// 	return nil
// case f.IsContainer:
// 	return f.resolveContainer(l, o)
// case !f.IsTagged:
// 	return nil
// }
// var required bool
// var ignoreCase bool
// var defaultValue string

// if f.Tag.IgnoreCaseIsSet {
// 	ignoreCase = f.Tag.IgnoreCase
// } else {
// 	ignoreCase = o.IgnoreCase
// }

// if f.Tag.DefaultIsSet {
// 	defaultValue = f.Tag.Default
// } else {
// 	defaultValue = o.Default
// }

// if f.Tag.RequiredIsSet {
// 	required = f.Tag.Required
// } else {
// 	required = o.RequireAllFields
// }
