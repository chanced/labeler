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
	path        string
	Key         string
	WasSet      bool
	Keep        bool
	IsTagged    bool
	IsContainer bool
}

func newField(parent reflected, i int, o Options) (*field, error) {
	sf, ok := parent.StructField(i)
	if !ok {
		panic(errors.New("can not access field"))
	}
	rv, ok := parent.ValueField(i)
	if !ok {
		panic(errors.New("can not access field"))
	}

	fieldName := sf.Name
	f := &field{
		Name:   fieldName,
		Parent: parent,
	}
	tag, err := f.parseTag(sf, o)
	if err != nil {
		return f, f.err(err)
	}

	f.Tag = tag
	if tag != nil {
		f.Key = tag.Key
	}

	f.setIsContainer(o)

	f.meta = newMeta(rv)

	if err != nil {
		return f, f.err(err)
	}
	if !f.canAddr && f.IsTagged {
		return f, f.err(ErrUnexportedField)
	}

	if f.IsTagged && !f.canAddr {
		return f, f.err(ErrUnexportedField)
	}

	if f.IsTagged || f.IsContainer {
		f.unmarshal = getUnmarshal(f, o)
		f.marshal = getMarshal(f, o)
		if f.unmarshal == nil {
			return f, f.err(ErrUnsupportedType)
		}
		if f.marshal == nil {
			return f, f.err(ErrUnsupportedType)
		}
	}

	return f, nil
}

func (f *field) Unmarshal(kvs *keyvalues, o Options) error {
	if f.unmarshal == nil {
		// this shouldn't happen. just being safe.
		return f.err(ErrUnsupportedType)
	}
	return f.unmarshal(f, kvs, o)
}
func (f *field) Marshal(kvs *keyvalues, o Options) error {
	if f.marshal == nil {
		// this shouldn't happen. just being safe.
		return f.err(ErrUnsupportedType)
	}
	return f.marshal(f, kvs, o)
}

func (f *field) ignoreCase(o Options) bool {
	if f.Tag.IgnoreCaseIsSet {
		return f.Tag.IgnoreCase
	}
	return o.IgnoreCase
}

func (f *field) parseTag(sf reflect.StructField, o Options) (*Tag, error) {
	tagstr, isTagged := sf.Tag.Lookup(o.Tag)
	f.IsTagged = isTagged
	if !isTagged {
		return nil, nil
	}
	return newTag(tagstr, o)
}

func (f *field) setIsContainer(o Options) {
	switch {
	case f.Tag != nil && f.Tag.IsContainer:
		f.IsContainer = true
	case o.ContainerField != "" && o.ContainerField == f.path:
		f.IsContainer = true
	}
}

func (f *field) Path() string {
	if f.path != "" {
		return f.path
	}
	if f.Parent.Path() != "" {
		return fmt.Sprintf("%s.%s", f.Parent.Path(), f.Name)
	}
	return f.Name
}

func (f *field) err(err error) *FieldError {
	if err != nil {
		return newFieldError(f, err)
	}
	return nil
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

func (f *field) formatInt(o Options) (string, error) {
	switch f.kind {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		v := f.value.Int()
		return strconv.FormatInt(v, f.intBase(o)), nil
	default:
		return "", nil
	}
}

func (f *field) setInt(s string, bits int, o Options) error {
	v, err := strconv.ParseInt(s, f.intBase(o), bits)
	if err != nil {
		return f.err(err)
	}
	f.value.SetInt(v)
	return nil
}

func (f *field) formatString(o Options) (string, error) {
	return f.value.String(), nil
}

func (f *field) setString(s string, o Options) error {
	f.value.SetString(s)
	return nil
}

func (f *field) formatUint(o Options) (string, error) {
	switch f.kind {
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		v := f.value.Uint()
		return strconv.FormatUint(v, f.intBase(o)), nil
	default:
		return "", nil
	}
}

func (f *field) setUint(s string, bits int, o Options) error {
	v, err := strconv.ParseUint(s, f.uintBase(o), bits)
	if err != nil {
		return f.err(err)
	}
	f.value.SetUint(v)
	return nil
}

func (f *field) formatComplex(o Options) (string, error) {
	switch f.kind {
	case reflect.Complex128:
		v := f.value.Complex()
		return strconv.FormatComplex(v, f.complexFormat(o), -1, 128), nil
	case reflect.Complex64:
		v := f.value.Complex()
		return strconv.FormatComplex(v, f.complexFormat(o), -1, 64), nil
	default:
		return "", nil
	}
}

func (f *field) setComplex(s string, bits int, o Options) error {
	v, err := strconv.ParseComplex(s, bits)
	if err != nil {
		return f.err(err)
	}
	f.value.SetComplex(v)
	return nil
}

func (f *field) formatFloat(o Options) (string, error) {
	switch f.kind {
	case reflect.Float64:
		v := f.value.Float()
		return strconv.FormatFloat(v, f.floatFormat(o), -1, 64), nil
	case reflect.Float32:
		v := f.value.Float()
		return strconv.FormatFloat(v, f.floatFormat(o), -1, 32), nil
	default:
		return "", nil
	}
}

func (f *field) setFloat(s string, bits int, o Options) error {
	v, err := strconv.ParseFloat(s, bits)
	if err != nil {
		return f.err(err)
	}
	f.value.SetFloat(v)
	return nil
}

func (f *field) formatBool(o Options) (string, error) {
	if f.kind == reflect.Bool {
		v := f.value.Bool()
		return strconv.FormatBool(v), nil
	}
	return "", nil
}

func (f *field) setBool(s string, o Options) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return f.err(err)
	}
	f.value.SetBool(v)
	return nil
}

func (f *field) formatTime(o Options) (string, error) {
	if v, ok := f.Interface().(time.Time); ok {
		return v.Format(f.timeFormat(o)), nil
	}
	return "", nil
}

func (f *field) setTime(s string, o Options) error {
	if !timeType.AssignableTo(f.Type()) {
		return f.err(errors.New("Can not assign time.Time to " + f.Name))
	}
	v, err := time.Parse(f.timeFormat(o), s)
	if err != nil {
		return f.err(err)
	}
	rv := reflect.ValueOf(v)
	f.value.Set(rv)
	return nil
}

func (f *field) formatDuration(o Options) (string, error) {
	if v, ok := f.Interface().(time.Duration); ok {
		return v.String(), nil
	}
	return "", nil

}

func (f *field) setDuration(s string, o Options) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return f.err(err)
	}
	rv := reflect.ValueOf(v)
	f.value.Set(rv)
	return nil
}

func (f *field) setMap(v map[string]string, o Options) error {
	if f.kind != reflect.Map {
		return f.err(errors.New("invalid type")) // this shouldn't happen
	}
	f.value.Set(reflect.ValueOf(v))
	return nil
}

func (f *field) ShouldKeep(o Options) bool {
	if f.Tag.KeepIsSet {
		return f.Tag.Keep
	}
	return o.KeepLabels
}

func (f *field) ShouldDiscard(o Options) bool {
	return !f.ShouldKeep(o)
}

func (f *field) Default(o Options) string {
	if f.Tag.DefaultIsSet {
		return f.Tag.Default
	}
	return o.Default
}

func (f *field) Save() {
	f.save()
	f.Parent.Save()
}

func (f *field) Topic() topic {
	return fieldTopic
}

func (f *field) IsFieldContainer() bool {
	return f.IsContainer
}
