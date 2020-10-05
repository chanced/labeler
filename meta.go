package labeler

import (
	"reflect"
)

type reflected interface {
	Meta() meta
	Save()
	topic() topic
	path() string
	numField() int
	implements(u reflect.Type) bool
	assignableTo(u reflect.Type) bool
}

type topic int

const (
	invalidTopic = iota
	fieldTopic
	subjectTopic
	inputTopic
)

type meta struct {
	Type          reflect.Type
	Kind          reflect.Kind
	Value         reflect.Value
	Field         reflect.StructField
	Addr          reflect.Value
	Ptr           reflect.Value
	Interface     interface{}
	TypeName      string
	PkgPath       string
	NumField      int
	IsPtr         bool
	CanAddr       bool
	CanSet        bool
	CanInterface  bool
	IsStructField bool
}

func newMeta(rv reflect.Value) meta {
	kind := rv.Kind()
	t := rv.Type()
	tname := t.Name()
	pkgPath := t.PkgPath()

	m := meta{
		Value:        rv,
		Kind:         kind,
		Type:         t,
		IsPtr:        kind == reflect.Ptr,
		CanAddr:      rv.CanAddr(),
		CanInterface: rv.CanInterface(),
		TypeName:     tname,
		PkgPath:      pkgPath,
	}

	m.IsPtr = m.deref()

	if m.Kind == reflect.Struct {
		m.NumField = m.Type.NumField()
	}

	if m.CanInterface {
		m.Interface = rv.Interface()
	}

	return m
}

func (m *meta) deref() bool {
	if m.Kind != reflect.Ptr {
		return false
	}
	m.IsPtr = true
	var ptr reflect.Value
	if m.Value.IsNil() {
		elem := m.Type.Elem()
		ptr = reflect.New(elem)
	} else {
		ptr = m.Value
	}
	m.Ptr = m.Value
	m.Value = ptr.Elem()
	m.Type = ptr.Type()
	m.Kind = ptr.Kind()
	m.CanAddr = m.Value.CanAddr()
	return true

}

func (m *meta) IsStruct() bool {
	return m.Kind == reflect.Struct
}

func (m *meta) save() {
	if m.IsPtr {
		m.Ptr.Set(m.Value)
	}
}

func (m meta) Meta() meta {
	return m
}

func (m meta) numField() int {
	return m.NumField
}

func (m meta) implements(u reflect.Type) bool {
	return m.Type.Implements(u)
}

func (m meta) assignableTo(u reflect.Type) bool {
	return m.Type.AssignableTo(u)
}
