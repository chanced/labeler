package labeler

import (
	"reflect"
)

type reflected interface {
	Meta() meta
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
	Kind() reflect.Kind
	IsFieldContainer() bool
	Unmarshal(kvs *keyvalues, o Options) error
	Marshal(kvs *keyvalues, o Options) error
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
	rtype        reflect.Type
	kind         reflect.Kind
	value        reflect.Value
	field        reflect.StructField
	addr         reflect.Value
	addrType     reflect.Type
	ptrValue     reflect.Value
	ptrType      reflect.Type
	typeName     string
	pkgPath      string
	numField     int
	isPtr        bool
	canAddr      bool
	canSet       bool
	canInterface bool
	marshal      marshal
	unmarshal    unmarshal
	// unmarshaler  unmarshaler
}

func newMeta(rv reflect.Value) meta {
	m := meta{
		value: rv,
		kind:  rv.Kind(),
		rtype: rv.Type(),
	}
	m.canSet = m.value.CanSet()
	m.canAddr = m.value.CanAddr()
	m.canInterface = m.value.CanInterface()

	m.isPtr = m.deref()
	m.typeName = m.rtype.Name()
	m.pkgPath = m.rtype.PkgPath()
	if m.canAddr {
		m.addr = m.value.Addr()
		m.addrType = m.addr.Type()
	}
	if m.kind == reflect.Struct {
		m.numField = m.rtype.NumField()
	}

	return m
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
	return m.rtype
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

func (m *meta) Kind() reflect.Kind {
	return m.kind
}

func (m *meta) deref() bool {
	if m.kind != reflect.Ptr {
		return false
	}
	var ptr reflect.Value
	if m.value.IsNil() {
		elem := m.rtype.Elem()
		ptr = reflect.New(elem).Elem()
	} else {
		ptr = m.value.Elem()
	}
	m.ptrValue = m.value
	m.ptrType = m.rtype
	m.value = ptr
	m.rtype = ptr.Type()
	m.kind = ptr.Kind()
	// m.canAddr = m.value.CanAddr()
	// if m.kind == reflect.Ptr && !m.isPtrPtr {
	// 	m.ptrPtrValue = m.ptrValue
	// 	m.isPtrPtr = true
	// 	return m.deref()
	// }
	return true

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

func (m meta) Meta() meta {
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
	if m.rtype.Implements(u) {
		return true
	}

	if !m.canAddr {
		return false
	}
	return m.addrType.Implements(u)
}

func (m meta) Assignable(u reflect.Type) bool {
	return u.AssignableTo(m.rtype)
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
	if m.kind != reflect.Struct {
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
	return m.rtype.Field(i), true
}
