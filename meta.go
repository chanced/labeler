package labeler

import (
	"reflect"
)

type reflected interface {
	Meta() meta
	Save()
	topic() topic
	path() string
	NumField() int
	Implements(u reflect.Type) bool
	Assignable(u reflect.Type) bool
	CanInterface() bool
	CanAddr() bool
	CanSet() bool
}

type topic int

const (
	invalidTopic = iota
	fieldTopic
	subjectTopic
	inputTopic
)

//TODO: rename the values below and create interface accessors where applicable
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
	numField      int
	IsPtr         bool
	canAddr       bool
	canSet        bool
	canInterface  bool
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
		canAddr:      rv.CanAddr(),
		canSet:       rv.CanSet(),
		canInterface: rv.CanInterface(),
		TypeName:     tname,
		PkgPath:      pkgPath,
	}

	m.IsPtr = m.deref()
	m.deref()
	if m.Kind == reflect.Struct {
		m.numField = m.Type.NumField()
	}

	if m.canInterface {
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
	m.canAddr = m.Value.CanAddr()
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

func (m meta) NumField() int {
	return m.numField
}

func (m meta) Implements(u reflect.Type) bool {
	return m.Type.Implements(u)
}

func (m meta) Assignable(u reflect.Type) bool {
	return u.AssignableTo(m.Type)
}

func (m meta) CanSet() bool {
	return m.canSet
}

func (m meta) CanInterface() bool {
	return m.canInterface
}
func (m meta) CanAddr() bool {
	return m.canAddr
}
