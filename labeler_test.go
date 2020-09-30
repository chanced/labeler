package labeler

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type StructWithLabels struct {
	Labels map[string]string
}

func (s StructWithLabels) GetLabels() map[string]string {
	return s.Labels
}
func (s StructWithLabels) SetLabels(l map[string]string) {
	s.Labels = l
}

type MyEnum int

const (
	EnumUnknown MyEnum = iota
	EnumValA
	EnumValB
)

var myEnumMapToStr map[MyEnum]string = map[MyEnum]string{
	EnumUnknown: "Unknown",
	EnumValA:    "ValueA",
	EnumValB:    "ValueB",
}

func getMyEnumMapFromStr() map[string]MyEnum {
	m := make(map[string]MyEnum)
	for key, value := range myEnumMapToStr {
		m[value] = key
	}
	return m
}

var myEnumMapFromStr map[string]MyEnum = getMyEnumMapFromStr()

func (my MyEnum) String() string {
	return myEnumMapToStr[my]
}

func (my *MyEnum) FromString(s string) error {
	if v, ok := myEnumMapFromStr[s]; ok {
		*my = v
		return nil
	}
	return errors.New("Invalid value")
}

type Example struct {
	Name            string        `label:"name"`
	Important       string        `label:"imp, required"`
	Enum            MyEnum        `label:"enum"`
	Duration        time.Duration `label:"duration"`
	Time            time.Time     `label:"time, format: 01/02/2006 03:04PM"`
	Dedupe          string        `label:"dedupe, discard"`
	CaSe            string        `label:"CaSe, casesensitive"`
	FloatWithFormat float64       `label:"FloatWithFormat, format:b"`
	Float64         float64       `label:"float64"`
	Float32         float32       `label:"float32"`
	Int             int           `label:"int"`
	Int64           int64         `label:"int64"`
	Int32           int32         `label:"int32"`
	Int16           int16         `label:"int16"`
	Int8            int8          `label:"int8"`
	Bool            bool          `label:"bool"`
	Uint            uint          `label:"uint"`
	Uint64          uint64        `label:"uint64"`
	Uint32          uint32        `label:"uint32"`
	Uint16          uint16        `label:"uint16"`
	Uint8           uint8         `label:"uint8"`

	Labels map[string]string
}

func (e *Example) SetLabels(l map[string]string) {
	e.Labels = l
}

func (e *Example) GetLabels() map[string]string {
	return e.Labels
}

func TestExample(t *testing.T) {

	labels := map[string]string{
		"name":            "Archer",
		"imp":             "important field",
		"enum":            "ValueB",
		"int":             "123456789",
		"int64":           "1234567890",
		"int32":           "12345",
		"int16":           "123",
		"int8":            "1",
		"bool":            "true",
		"duration":        "1s",
		"float64":         "1.1234567890",
		"float32":         "1.123",
		"time":            "09/26/2020 10:10PM",
		"uint":            "1234",
		"uint64":          "1234567890",
		"uint32":          "1234567",
		"uint16":          "123",
		"uint8":           "1",
		"FloatWithFormat": "123.234823484",
		"dedupe":          "Demonstrates that discard is removed from the Labels after field value is set",
		"case":            "value should not be set due to not matching case",
	}

	l := StructWithLabels{
		Labels: labels,
	}

	v := &Example{}
	err := Unmarshal(l, v)
	assert.NoError(t, err, "Should not have thrown an error")

	assert.Equal(t, "Archer", v.Name, "Name should be set to \"Archer\"")
	assert.Equal(t, EnumValB, v.Enum, "Enum should be set to EnumValB")
	assert.Equal(t, true, v.Bool, "Bool should be set to true")
	assert.Equal(t, 123456789, v.Int, "Int should be set to 123456789")
	assert.Equal(t, int8(1), v.Int8, "Int8 should be set to 1")
	assert.Equal(t, int16(123), v.Int16, "Int16 should be set to 123")
	assert.Equal(t, int32(12345), v.Int32, "Int32 should be set to 12345")
	assert.Equal(t, int64(1234567890), v.Int64, "Int64 should be set to 1234567890")
	assert.Equal(t, float64(1.1234567890), v.Float64, "Float64 should be ste to 1.1234567890")
	assert.Equal(t, float32(1.123), v.Float32, "Float32 should be 1.123")
	assert.Equal(t, time.Second*1, v.Duration, "Duration should be 1 second")
	assert.Equal(t, uint(1234), v.Uint, "Unit should be set to 1234")
	assert.Equal(t, uint64(1234567890), v.Uint64, "Uint64 should be set to 1234567890")
	assert.Equal(t, uint32(1234567), v.Uint32, "Uinit32 should be set to 1234567")
	assert.Equal(t, uint16(123), v.Uint16, "Unit16 should be set to 123")
	assert.Equal(t, uint8(1), v.Uint8, "Uint8 should be set to 1")

	assert.Zero(t, v.CaSe)
	assert.Equal(t, "Demonstrates that discard is removed from the Labels after field value is set", v.Dedupe)
	assert.NotContains(t, v.GetLabels(), "dedupe")
	assert.Equal(t, time.Date(int(2020), time.September, int(26), int(22), int(10), int(0), int(0), time.UTC), v.Time)

	// res, err := Marshal(v)

	// for key, value := range labels {
	// 	assert.Contains(t, res, key)
	// 	assert.Equal(t, value, res[key])
	// }
}

type InvalidDueToMissingLabels struct {
	Name string `label:"name,required"`
	Enum MyEnum `label:"enum"`
}

func TestInvalidDueToMissingLabels(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"enum": "ValueB",
		},
	}
	inv := &InvalidDueToMissingLabels{}
	err := Unmarshal(l, inv)
	assert.Error(t, err, "Should have thrown an error")
	assert.Error(t, ErrInvalidValue, err)
	t.Log(err)
}

type WithValidation struct {
	Name          string            `label:"name"`
	Enum          MyEnum            `label:"enum,required"`
	RequiredField string            `label:"required_field,required"`
	Defaulted     string            `label:"defaulted,default:default value"`
	Labels        map[string]string `label:"*"`
}

func TestLabelerWithValidation(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"name": "my name",
			"enum": "X",
		},
	}

	v := &WithValidation{}
	err := Unmarshal(l, v)
	assert.Error(t, err, "should contain errors")
	var e *ParsingError
	if errors.As(err, &e) {
		assert.Len(t, e.Errors, 2)
	} else {
		assert.Fail(t, "error should be a parsing error")
	}
	assert.Equal(t, "my name", v.Name)
	assert.Equal(t, EnumUnknown, v.Enum)
}

type WithDiscard struct {
	Discarded string `label:"will_not_be_in_labels,discard"`
	Kept      string `label:"will_be_in_labels"`
	labels    map[string]string
}

func (wd *WithDiscard) SetLabels(labels map[string]string) {
	wd.labels = labels
}

func TestLabelerWithDiscard(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"will_not_be_in_labels": "discarded_value",
			"will_be_in_labels":     "kept_value",
			"unassigned":            "unassigned will be in labels",
		},
	}

	v := &WithDiscard{}
	err := Unmarshal(l, v)
	assert.NoError(t, err)
	assert.Equal(t, "discarded_value", v.Discarded)
	assert.Equal(t, "kept_value", v.Kept)
	assert.NotContains(t, v.labels, "will_not_be_in_labels")
	assert.Contains(t, v.labels, "will_be_in_labels")
	assert.Contains(t, v.labels, "unassigned")
}

type NestedWithRequired struct {
	SubField string `label:"subfield,required"`
}

type WithNestedRequired struct {
	Nested      NestedWithRequired
	ParentField string            `label:"parentfield"`
	Labels      map[string]string `label:"*"`
}

func TestLabelerWithNestedStruct(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"parentfield": "parent-value",
			"subfield":    "sub-value",
		},
	}

	v := &WithNestedRequired{}
	err := Unmarshal(l, v)
	assert.NoError(t, err)
	assert.Equal(t, "sub-value", v.Nested.SubField)
}

type WithNestedStructAsPtr struct {
	Nested *NestedWithRequired
}

func (p *WithNestedStructAsPtr) SetLabels(m map[string]string) {

}
func TestLabelerWithNestedStructAsPtr(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"parentfield": "parent-value",
			"subfield":    "sub-value",
		},
	}

	v := &WithNestedStructAsPtr{}
	err := Unmarshal(l, v)
	var p *ParsingError
	if errors.As(err, &p) {
		t.Log(p.Errors)
	}
	assert.NoError(t, err)
	assert.Equal(t, "sub-value", v.Nested.SubField)

}

func TestFieldPanicRecovery(t *testing.T) {

}

func TestLabelerInvalidWithNestedStruct(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"parentfield": "parent-value",
		},
	}
	v := &WithNestedRequired{}
	err := Unmarshal(l, v)
	assert.Error(t, err)

}
