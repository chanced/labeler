package labeler

import (
	"fmt"
	"reflect"
)

type meta struct {
	Type          reflect.Type
	Kind          reflect.Kind
	Value         reflect.Value
	Field         reflect.StructField
	Addr          reflect.Value
	Ptr           reflect.Value
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
		Value:    rv,
		Kind:     kind,
		Type:     t,
		IsPtr:    kind == reflect.Ptr,
		CanAddr:  rv.CanAddr(),
		TypeName: tname,
		PkgPath:  pkgPath,
	}
	m.IsPtr = m.deref()
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

func (m *meta) SetStructDetails() {
	if m.Kind == reflect.Struct {
		m.NumField = m.Type.NumField()
	}
}

func (m *meta) IsStruct() bool {
	return m.Kind == reflect.Struct
}

func (m meta) Meta() meta {
	return m
}

func (m *meta) GetTypePath() string {
	return fmt.Sprintf("%s.%s", m.PkgPath, m.TypeName)
}
