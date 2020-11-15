package labeler

import (
	"reflect"
)

type reflected interface {
	Meta() *meta
	Save()
	Topic() topic
	Path() string
	Type() reflect.Type
	IsStruct() bool
	NumField() int
	Value() reflect.Value
	Implements(u reflect.Type) bool
	Interface() interface{}
	Assignable(u reflect.Type) bool
	CanInterface() bool
	CanAddr() bool
	CanSet() bool
	StructField(i int) (reflect.StructField, bool)
	ValueField(i int) (reflect.Value, bool)
	StoreValue(reflect.Value)
	TypeName() string
	PkgPath() string
	IsPtr() bool
	PtrValue() reflect.Value
	Kind() reflect.Kind
	Unmarshal(kvs *keyValues, o Options) error
	Marshal(kvs *keyValues, o Options) error
	IsContainer(o Options) bool
	ColType() reflect.Type
	SetValue(reflect.Value)
	IsArray() bool
	IsSlice() bool
	Len() int
	ColValue() reflect.Value
	IsElem() bool
	SetIsElem(bool)
	deref() bool
	ResetCollection()
	ColElemKind() reflect.Kind
	ColElemType() reflect.Type
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
	typ          reflect.Type
	kind         reflect.Kind
	value        reflect.Value
	field        reflect.StructField
	addr         reflect.Value
	ptrType      reflect.Type
	colValue     reflect.Value
	colType      reflect.Type
	colKind      reflect.Kind
	colElemType  reflect.Type
	colElemKind  reflect.Kind
	addrType     reflect.Type
	ptrValue     reflect.Value
	typeName     string
	pkgPath      string
	numField     int
	isPtr        bool
	isArray      bool
	isSlice      bool
	canAddr      bool
	canSet       bool
	canInterface bool
	len          int
	isElem       bool
	marshal      marshalFunc
	unmarshal    unmarshalFunc
	// unmarshaler  unmarshaler
}

func newMeta(rv reflect.Value) meta {
	m := meta{
		value: rv,
		kind:  rv.Kind(),
		typ:   rv.Type(),
	}
	m.canSet = m.value.CanSet()
	m.canAddr = m.value.CanAddr()
	m.canInterface = m.value.CanInterface()
	m.checkArraySlice()
	m.isPtr = m.deref()

	m.typeName = m.typ.Name()
	m.pkgPath = m.typ.PkgPath()
	if m.canAddr {
		m.addr = m.value.Addr()
		m.addrType = m.addr.Type()
	}
	if m.kind == reflect.Struct {
		m.numField = m.typ.NumField()
	}

	return m
}
func (m *meta) ColElemKind() reflect.Kind {
	return m.colElemKind
}

func (m *meta) ColElemType() reflect.Type {
	return m.colElemType
}

func (m *meta) Interface() interface{} {
	if m.isPtr {
		return m.ptrValue.Interface()
	}
	if m.value.CanAddr() && !m.addr.IsZero() {
		return m.addr.Interface()
	}
	return m.value.Interface()
}

func (m *meta) Type() reflect.Type {
	return m.typ
}

func (m *meta) ColType() reflect.Type {
	return m.colType
}

func (m *meta) IsElem() bool {
	return m.isElem
}

func (m *meta) SetIsElem(v bool) {
	m.isElem = v
}

func (m *meta) StoreValue(v reflect.Value) {
	m.value = v
}

func (m *meta) PkgPath() string {
	return m.pkgPath
}

func (m *meta) TypeName() string {
	return m.typeName
}

func (m *meta) ColValue() reflect.Value {
	return m.colValue
}
func (m *meta) IsSlice() bool {
	return m.isSlice
}

func (m *meta) IsArray() bool {
	return m.isArray
}

func (m *meta) Len() int {
	return m.len
}
func (m *meta) Kind() reflect.Kind {
	return m.kind
}

func (m *meta) checkArraySlice() bool {

	if m.kind != reflect.Slice && m.kind != reflect.Array {
		return false
	}

	m.len = m.value.Len()
	m.isSlice = m.kind == reflect.Slice
	m.isArray = m.kind == reflect.Array

	if m.isSlice && m.value.IsNil() {
		m.value.Set(reflect.New(m.typ).Elem())
	}

	m.colType = m.typ
	m.colValue = m.value
	m.colKind = m.kind
	m.value = reflect.New(m.typ).Elem()
	m.typ = m.typ.Elem()
	m.kind = m.typ.Kind()
	m.colElemKind = m.kind
	m.colElemType = m.typ
	return true
}

func (m *meta) ResetCollection() {
	if !m.isArray && !m.isSlice {
		return
	}
	m.typ = m.colType
	m.kind = m.colKind
	m.value = m.colValue

}

func (m *meta) deref() bool {
	if m.kind != reflect.Ptr {
		return false
	}
	var ptr reflect.Value
	if m.value.IsNil() {
		elem := m.typ.Elem()
		ptr = reflect.New(elem).Elem()
	} else {
		ptr = m.value.Elem()
	}
	m.ptrValue = m.value
	m.ptrType = m.typ
	m.value = ptr
	m.typ = ptr.Type()
	m.kind = ptr.Kind()

	return true

}

func (m *meta) SetValue(rv reflect.Value) {
	m.value = rv
}

func (m *meta) IsStruct() bool {
	return m.kind == reflect.Struct
}

func (m *meta) save() {

	if m.isPtr && m.CanSet() {
		m.ptrValue.Set(m.value.Addr())
	}
}

func (m *meta) IsPtr() bool {
	return m.isPtr
}

func (m *meta) PtrValue() reflect.Value {
	return m.ptrValue
}

func (m *meta) Meta() *meta {
	return m
}

func (m meta) NumField() int {
	return m.numField
}

func (m meta) Implements(u reflect.Type) bool {
	name := m.typeName
	_ = name
	uname := u.Name()
	_ = uname
	if m.isPtr && m.ptrType.Implements(u) {
		return true
	}
	if m.typ.Implements(u) {
		return true
	}

	if !m.canAddr {
		return false
	}
	return m.addrType.Implements(u)
}

func (m meta) Assignable(u reflect.Type) bool {
	return u.AssignableTo(m.typ)
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

func (m meta) Value() reflect.Value {
	return m.value
}

func (m meta) ValueField(i int) (reflect.Value, bool) {
	if m.kind != reflect.Struct && m.value.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	if i >= m.numField {
		return reflect.Value{}, false
	}
	return m.value.Field(i), true
}
func (m meta) StructField(i int) (reflect.StructField, bool) {
	if m.kind != reflect.Struct {
		return reflect.StructField{}, false
	}
	if i >= m.numField {
		return reflect.StructField{}, false
	}
	return m.typ.Field(i), true
}
