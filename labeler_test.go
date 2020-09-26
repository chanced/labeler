package labeler

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type StructWithLabels struct {
	Labels map[string]string
}

func (l StructWithLabels) GetLabels() map[string]string {
	return l.Labels
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

type ValidBasic struct {
	Name     string        `label:"name"`
	Enum     MyEnum        `label:"enum"`
	Float64  float64       `label:"float64"`
	Float32  float32       `label:"float32"`
	Int      int           `label:"int"`
	Int64    int64         `label:"int64"`
	Int32    int32         `label:"int32"`
	Int16    int16         `label:"int16"`
	Int8     int8          `label:"int8"`
	Bool     bool          `label:"bool"`
	Duration time.Duration `label:"duration"`
	Time     time.Time     `label:"time,layout:01/02/2006 03:04PM"`
	Labels   map[string]string
}

func (v *ValidBasic) SetLabels(l map[string]string) {
	v.Labels = l

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

func TestValidBasic(t *testing.T) {

	l := StructWithLabels{
		Labels: map[string]string{
			"name":     "my name",
			"enum":     "ValueB",
			"int":      "123456789",
			"int64":    "1234567890000",
			"int32":    "12345",
			"int16":    "123",
			"int8":     "1",
			"bool":     "true",
			"duration": "1s",
			"float64":  "1.1234567890",
			"float32":  "1.123",
			"time":     "09/26/2020 10:10PM",
		},
	}

	v := &ValidBasic{}
	err := Unmarshal(l, v)
	assert.NoError(t, err, "Should not have thrown an error")

	assert.Equal(t, "my name", v.Name)
	assert.Equal(t, EnumValB, v.Enum)
	assert.Equal(t, true, v.Bool)
	assert.Equal(t, 123456789, v.Int)
	assert.Equal(t, int8(1), v.Int8)
	assert.Equal(t, int16(123), v.Int16)
	assert.Equal(t, int32(12345), v.Int32)
	assert.Equal(t, int64(1234567890000), v.Int64)
	assert.Equal(t, float64(1.1234567890), v.Float64)
	assert.Equal(t, float32(1.123), v.Float32)
	assert.Equal(t, time.Second*1, v.Duration)

	t.Log("TIME", v.Time)

	assert.Equal(t, time.Date(int(2020), time.September, int(26), int(22), int(10), int(0), int(0), time.UTC), v.Time)
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
		assert.Len(t, e.Errs, 2)
	} else {
		assert.Fail(t, "Error should be a parsing error")
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

type NestedStruct struct {
	SubField string `label:"subfield,required"`
}

type WithNestedStruct struct {
	Nested      NestedStruct
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

	v := &WithNestedStruct{}
	err := Unmarshal(l, v)
	assert.NoError(t, err)
	assert.Equal(t, "sub-value", v.Nested.SubField)
}

type WithNestedStructAsPtr struct {
	Nested *NestedStruct
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

		fmt.Println(p.Errs)
		t.Log(p.Errs)
	}
	assert.NoError(t, err)
	assert.Equal(t, "sub-value", v.Nested.SubField)

}

func TestLabelerWithNestedStructWithValidationErrs(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"parentfield": "parent-value",
		},
	}

	v := &WithNestedStruct{}
	err := Unmarshal(l, v)
	assert.Error(t, err)

}
