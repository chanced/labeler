package labeler

import (
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

func (my *MyEnum) FromString(s string) bool {
	if v, ok := myEnumMapFromStr[s]; ok {
		*my = v
		return true
	}
	return false
}

type InvalidDueTomissingLabels struct {
	Name string `label:"name"`
	Enum MyEnum `label:"enum"`
}

type ValidBasicLabeler struct {
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

	Labels map[string]string
}

func (v *ValidBasicLabeler) SetLabels(l map[string]string) {
	v.Labels = l

}
func TestInvalidDueToMissingLabels(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"name": "my name",
			"enum": "ValueB",
		},
	}
	inv := &InvalidDueTomissingLabels{}
	err := Unmarshal(l, inv)
	assert.Error(t, err, "Should have thrown an error")
	t.Log(err)
}
func TestValidBasicLabeler(t *testing.T) {

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
		},
	}

	v := &ValidBasicLabeler{}
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
	t.Log(v)
}

func TestValidBasicLabeler(t *testing.T) {

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
		},
	}

	v := &ValidBasicLabeler{}
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
	t.Log(v)
}
