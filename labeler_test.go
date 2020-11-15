package labeler

import (
	"errors"
	"fmt"
	"strings"
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
	for k, v := range myEnumMapToStr {
		m[v] = k
	}
	return m
}

var myEnumMapFromStr map[string]MyEnum = getMyEnumMapFromStr()

func (my *MyEnum) String() string {
	return myEnumMapToStr[*my]
}

var ErrExampleInvalidEnum = errors.New("invalid MyEnum Value")

func (my *MyEnum) FromString(s string) error {
	if v, ok := myEnumMapFromStr[s]; ok {
		*my = v
		return nil
	}
	return ErrExampleInvalidEnum
}

type Example struct {
	Name               string        `label:"name"`
	Enum               MyEnum        `label:"enum"`
	Duration           time.Duration `label:"duration"`
	Time               time.Time     `label:"time, format: 01/02/2006 03:04PM"`
	Time2              time.Time     `label:"time2, timeformat: 01/02/2006 03:04PM"`
	Dedupe             string        `label:"dedupe, discard"`
	WithDefault        string        `label:"withdefault, default:defaultvalue"`
	CaSe               string        `label:"CaSe, casesensitive"`
	FloatWithFormat    float64       `label:"floatWithFormat, format:b"`
	FloatWithFormat2   float64       `label:"floatWithFormat2, floatformat:b"`
	Complex128         complex128    `label:"complex128"`
	Complex64          complex64     `label:"complex64"`
	ComplexWithFormat  complex64     `label:"complexWithFormat,format:b"`
	ComplexWithFormat2 complex64     `label:"complexWithFormat2,complexformat:b"`
	Float64            float64       `label:"float64"`
	Float32            float32       `label:"float32"`
	Int                int           `label:"int"`
	IntBinary          int           `label:"intbinary,base:2"`
	Int64              int64         `label:"int64"`
	Int32              int32         `label:"int32"`
	Int16              int16         `label:"int16"`
	Int8               int8          `label:"int8"`
	Bool               bool          `label:"bool"`
	Uint               uint          `label:"uint"`
	Uint64             uint64        `label:"uint64"`
	Uint32             uint32        `label:"uint32"`
	Uint16             uint16        `label:"uint16"`
	Uint8              uint8         `label:"uint8"`

	Labels map[string]string
}

func (e *Example) SetLabels(l map[string]string) {
	e.Labels = l
}

func (e *Example) GetLabels() map[string]string {
	return e.Labels
}

type ExampleWithEnum struct {
	Enum   MyEnum            `label:"enum"`
	Labels map[string]string `label:"*"`
}

func TestUnmarshalExample(t *testing.T) {
	labels := map[string]string{
		"name":               "Archer",
		"enum":               "ValueB",
		"int":                "123456789",
		"int64":              "1234567890",
		"int32":              "12345",
		"int16":              "123",
		"int8":               "1",
		"intbinary":          "111",
		"bool":               "true",
		"duration":           "1s",
		"float64":            "1.1234567890",
		"float32":            "1.123",
		"complex64":          "3+4i",
		"complex128":         "3+4i",
		"time":               "09/26/2020 10:10PM",
		"time2":              "09/26/2020 10:10PM",
		"uint":               "1234",
		"uint64":             "1234567890",
		"uint32":             "1234567",
		"uint16":             "123",
		"uint8":              "1",
		"floatWithFormat":    "123.234823484",
		"floatWithFormat2":   "123.234823484",
		"complexWithFormat":  "123.234823484",
		"complexWithFormat2": "123.234823484",
		"dedupe":             "Demonstrates that discard is removed from the Labels after field value is set",
		"case":               "value should not be set due to not matching case",
	}

	input := StructWithLabels{
		Labels: labels,
	}

	v := &Example{}
	err := Unmarshal(input, v)
	var pErr *ParsingError
	if errors.As(err, &pErr) {
		for _, e := range pErr.Errors {
			fmt.Println(e)
		}
	}
	assert.NoError(t, err, "Should not have thrown an error")

	assert.Equal(t, "Archer", v.Name, "Name should be set to \"Archer\"")
	assert.Equal(t, EnumValB, v.Enum, "Enum should be set to EnumValB")
	assert.Equal(t, true, v.Bool, "Bool should be set to true")
	assert.Equal(t, 123456789, v.Int, "Int should be set to 123456789")
	assert.Equal(t, int8(1), v.Int8, "Int8 should be set to 1")
	assert.Equal(t, int(7), v.IntBinary, "IntBinary should be set to 7")
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
	assert.Equal(t, "defaultvalue", v.WithDefault, "WithDefault should have been set to defaultvalue per tag")
	assert.Zero(t, v.CaSe)
	assert.Equal(t, "Demonstrates that discard is removed from the Labels after field value is set", v.Dedupe)
	assert.NotContains(t, v.GetLabels(), "dedupe")
	assert.Equal(t, time.Date(int(2020), time.September, int(26), int(22), int(10), int(0), int(0), time.UTC), v.Time)
	assert.Equal(t, time.Date(int(2020), time.September, int(26), int(22), int(10), int(0), int(0), time.UTC), v.Time2)
	fmt.Println(v)

}

func TestMarshalExample(t *testing.T) {
	labels := map[string]string{
		"name":               "Archer",
		"enum":               "ValueB",
		"int":                "123456789",
		"int64":              "1234567890",
		"int32":              "12345",
		"int16":              "123",
		"int8":               "1",
		"intbinary":          "111",
		"bool":               "true",
		"duration":           "1s",
		"float64":            "1.123456789",
		"float32":            "1.123",
		"complex64":          "(3+4i)",
		"complex128":         "(3+4i)",
		"time":               "09/26/2020 10:10PM",
		"time2":              "09/26/2020 10:10PM",
		"uint":               "1234",
		"uint64":             "1234567890",
		"uint32":             "12345",
		"uint16":             "123",
		"uint8":              "1",
		"floatWithFormat":    "8671879767525176p-46",
		"floatWithFormat2":   "8671879767525176p-46",
		"complexWithFormat":  "(16152635p-17+0p-149i)",
		"complexWithFormat2": "(16152635p-17+0p-149i)",
	}
	v := &Example{
		Name:               "Archer",
		Bool:               true,
		CaSe:               "",
		Duration:           1 * time.Second,
		Enum:               EnumValB,
		Complex128:         3 + 4i,
		Complex64:          3 + 4i,
		Float32:            1.123,
		Float64:            1.1234567890,
		Time:               time.Date(int(2020), time.September, int(26), int(22), int(10), int(0), int(0), time.UTC),
		Time2:              time.Date(int(2020), time.September, int(26), int(22), int(10), int(0), int(0), time.UTC),
		IntBinary:          7,
		Int:                123456789,
		Int64:              1234567890,
		Int32:              12345,
		Int16:              123,
		Int8:               1,
		Uint8:              1,
		Uint16:             123,
		Uint32:             12345,
		Uint64:             1234567890,
		Uint:               1234,
		FloatWithFormat:    123.234823484,
		FloatWithFormat2:   123.234823484,
		ComplexWithFormat:  123.234823484,
		ComplexWithFormat2: 123.234823484,
	}
	res, err := Marshal(v)
	assert.NoError(t, err)
	for key, value := range labels {
		fmt.Printf("%-20s %s%30s\n", key, res[key], value)
	}
	for key, value := range labels {
		assert.Contains(t, res, key, "marshaled results should contain ", key)
		v, ok := res[key]

		if ok {
			assert.Equal(t, v, value, fmt.Sprintf("%s should equal %v. got %s", key, value, v))
		}

	}

}

func TestInputAsMap(t *testing.T) {
	v := &Example{}
	labels := map[string]string{
		"name":               "Archer",
		"imp":                "important field",
		"enum":               "ValueB",
		"int":                "123456789",
		"int64":              "1234567890",
		"int32":              "12345",
		"int16":              "123",
		"int8":               "1",
		"intbinary":          "111",
		"bool":               "true",
		"duration":           "1s",
		"float64":            "1.1234567890",
		"float32":            "1.123",
		"complex64":          "(3+4i)",
		"complex128":         "(3+4i)",
		"time":               "09/26/2020 10:10PM",
		"time2":              "09/26/2020 10:10PM",
		"uint":               "1234",
		"uint64":             "1234567890",
		"uint32":             "1234567",
		"uint16":             "123",
		"uint8":              "1",
		"floatWithFormat":    "123.234823484",
		"floatWithFormat2":   "123.234823484",
		"complexWithFormat":  "123.234823484",
		"complexWithFormat2": "123.234823484",
		"dedupe":             "Demonstrates that discard is removed from the Labels after field value is set",
		"case":               "value should not be set due to not matching case",
	}

	err := Unmarshal(labels, v)
	assert.NoError(t, err, "Should not have thrown an error")

	assert.Equal(t, "Archer", v.Name, "Name should be set to \"Archer\"")
	assert.Equal(t, EnumValB, v.Enum, "Enum should be set to EnumValB")

}

func TestEnum(t *testing.T) {
	labels := map[string]string{
		"enum": "ValueB",
	}

	input := StructWithLabels{
		Labels: labels,
	}

	v := &ExampleWithEnum{}

	err := Unmarshal(input, v)
	assert.NoError(t, err, "Should not have thrown an error")
	assert.Equal(t, EnumValB, v.Enum, "Enum should be set to EnumValB")

}

type InvalidDueToNonaddressableContainer struct {
	Name   string            `label:"name"`
	labels map[string]string `label:"*"`
}

func TestInvalidValueDueToUnaccessibleContainer(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{},
	}

	v := &InvalidDueToNonaddressableContainer{}
	err := Unmarshal(l, v)
	assert.Error(t, err)
	var pErr *ParsingError
	if errors.As(err, &pErr) {
		if len(pErr.Errors) == 0 {
			assert.Fail(t, "ParsingError.Errors should contain an ErrUnexportedField for labels")
		} else {
			if !errors.Is(pErr.Errors[0], ErrUnexportedField) {
				assert.Fail(t, "ParsingError.Errors should contain an ErrUnexportedField for labels")
			}
		}

	} else {
		assert.Fail(t, "err should be a parsing error with")
	}
}

type WithDiscard struct {
	Discarded string `label:"will_not_be_in_labels,discard"`
	Kept      string `label:"will_be_in_labels"`
	Labels    map[string]string
}

func (wd *WithDiscard) SetLabels(labels map[string]string) {
	wd.Labels = labels
}

func TestLabeleeWithDiscard(t *testing.T) {
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
	assert.NotContains(t, v.Labels, "will_not_be_in_labels")
	assert.Contains(t, v.Labels, "will_be_in_labels")
	assert.Contains(t, v.Labels, "unassigned")
}

type Nested struct {
	SubField string `label:"subfield"`
}

type WithNested struct {
	Nested      Nested
	ParentField string            `label:"parentfield"`
	Labels      map[string]string `label:"*"`
}

func TestLabeleeWithNestedStruct(t *testing.T) {
	l := StructWithLabels{
		Labels: map[string]string{
			"parentfield": "parent-value",
			"subfield":    "sub-value",
		},
	}

	v := &WithNested{}
	err := Unmarshal(l, v)
	assert.NoError(t, err)
	assert.Equal(t, "sub-value", v.Nested.SubField)
}

type WithNestedStructAsPtr struct {
	Nested *Nested
}

func (p *WithNestedStructAsPtr) SetLabels(m map[string]string) {

}
func TestLabeleeWithNestedStructAsPtr(t *testing.T) {
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
	assert.NotNil(t, v.Nested)
	if v.Nested != nil {
		assert.Equal(t, "sub-value", v.Nested.SubField)
	}

}

type NumberBaseStruct struct {
	Labels      map[string]string `label:"*"`
	BinaryInt1  int               `label:"binaryint1, base:2"`
	BinaryInt2  int               `label:"binaryint2, intbase:2"`
	BinaryUint1 uint              `label:"binaryuint1, base:2"`
	BinaryUint2 uint              `label:"binaryuint2, uintbase:2"`
}

func TestBinaryNumbers(t *testing.T) {

	labels := map[string]string{
		"binaryInt1":  "111",
		"binaryInt2":  "11",
		"binaryUint1": "111",
		"binaryUint2": "11",
	}

	input := StructWithLabels{
		Labels: labels,
	}
	v := &NumberBaseStruct{}
	err := Unmarshal(input, v)
	assert.NoError(t, err, "Should not have thrown an error")
	assert.Equal(t, int(7), v.BinaryInt1, "BinaryInt1 should be set to 7")
	assert.Equal(t, int(3), v.BinaryInt2, "BinaryInt2 should be set to 3")
	assert.Equal(t, uint(7), v.BinaryUint1, "BinaryUint1 should be set to7")
	assert.Equal(t, uint(3), v.BinaryUint2, "BinaryUint2 should be set to 3")
}

type TestingSliceWithDefaultAndSplit struct {
	Slice  []int             `label:"strings,default:1|2|3,split:|"`
	Labels map[string]string `label:"*"`
}

func TestUnmarshalSliceWithDefault(t *testing.T) {
	v := new(TestingSliceWithDefaultAndSplit)
	in := make(map[string]string)

	err := Unmarshal(in, v)
	assert.NoError(t, err)
	assert.Contains(t, v.Slice, 1)
	assert.Contains(t, v.Slice, 2)
	assert.Contains(t, v.Slice, 3)

}

type TestingSlice struct {
	Strings []string `label:"strings"`
	Ints    []int    `label:"ints"`
	ex      string
	Labels  map[string]string `label:"*"`
}

func TestUnmarshalSlice(t *testing.T) {
	s := []string{"zero", "one", "two", "three", "four"}
	n := []string{"0", "1", "2", "3", "4", "2000"}
	ni := []int{0, 1, 2, 3, 4, 2000}

	sstr := strings.Join(s, ",")
	nstr := strings.Join(n, ",")
	m := map[string]string{"strings": sstr, "ints": nstr}
	var v TestingSlice
	err := Unmarshal(m, &v)
	assert.NoError(t, err)
	assert.Len(t, v.Strings, 5, "Slice")
	assert.Len(t, v.Ints, 6, "SliceInts")
	for i, sv := range s {
		assert.Equal(t, sv, v.Strings[i])
	}
	for i, nv := range ni {
		assert.Equal(t, nv, v.Ints[i])
	}
}

func TestMarshalSlice(t *testing.T) {
	s := []string{"zero", "one", "two", "three", "four"}
	n := []string{"0", "1", "2", "3", "4", "2000"}
	ni := []int{0, 1, 2, 3, 4, 2000}

	sstr := strings.Join(s, ",")
	istr := strings.Join(n, ",")

	v := TestingSlice{
		Strings: s,
		Ints:    ni,
	}

	mm, err := Marshal(&v)
	assert.NoError(t, err)
	assert.Equal(t, mm["strings"], sstr)
	assert.Equal(t, mm["ints"], istr)
	fmt.Println(mm)
}

type TestingArray struct {
	Strings [5]string         `label:"strings"`
	Floats  [6]float32        `label:"floats"`
	Labels  map[string]string `label:"*"`
}

func TestUnmarshalArray(t *testing.T) {
	a := [5]string{"zero", "one", "two", "three", "four"}
	n := []string{"0.1", "1.2", "2.3", "3.4", "4.5", "2000.0123"}
	nf := [6]float32{0.1, 1.2, 2.3, 3.4, 4.5, 2000.0123}

	astr := strings.Join(a[:], ",")
	nstr := strings.Join(n, ",")
	m := map[string]string{"Strings": astr, "floats": nstr}
	var v TestingArray
	err := Unmarshal(m, &v)
	assert.NoError(t, err)
	assert.Len(t, v.Strings, 5, "Strings")
	assert.Len(t, v.Floats, 6, "Floats")
	for i, sv := range a {
		assert.Equal(t, sv, v.Strings[i])
	}
	for i, nv := range nf {
		assert.Equal(t, nv, v.Floats[i])
	}

}

func TestMarshalArray(t *testing.T) {
	a := [5]string{"zero", "one", "two", "three", "four"}
	n := []string{"0.1", "1.2", "2.3", "3.4", "4.5", "2000.0123"}
	nf := [6]float32{0.1, 1.2, 2.3, 3.4, 4.5, 2000.0123}

	astr := strings.Join(a[:], ",")
	nstr := strings.Join(n, ",")
	v := TestingArray{
		Strings: a,
		Floats:  nf,
		Labels:  map[string]string{},
	}
	mm, err := Marshal(&v)

	assert.NoError(t, err)

	assert.Equal(t, astr, mm["strings"])
	assert.Equal(t, nstr, mm["floats"])
}

func TestOptionValidation(t *testing.T) {
	lbl := NewLabeler()
	err := lbl.ValidateOptions()
	assert.NoError(t, err)

	lbl = NewLabeler(OptTag(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid tag")

	lbl = NewLabeler(OptContainerToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid ContainerToken")

	lbl = NewLabeler(OptBaseToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid BaseToken")

	lbl = NewLabeler(OptIntBaseToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid IntBaseToken")

	lbl = NewLabeler(OptUintBaseToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid UintBaseToken")

	lbl = NewLabeler(OptDefaultToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid DefaultToken")

	lbl = NewLabeler(OptDiscardToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid DiscardToken")

	lbl = NewLabeler(OptFloatFormatToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid FloatFormatToken")

	lbl = NewLabeler(OptFormatToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid FormatToken")

	lbl = NewLabeler(OptComplexFormatToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid ComplexFormatToken")

	lbl = NewLabeler(OptIgnoreCaseToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid IgnoreCaseToken")

	lbl = NewLabeler(OptSeparator(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid Separator")

	lbl = NewLabeler(OptAssignmentStr(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid AssignmentStr")

	lbl = NewLabeler(OptTimeFormatToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid TimeFormatToken")

	lbl = NewLabeler(OptCaseSensitiveToken(""))
	err = lbl.ValidateOptions()
	assert.Error(t, err, "an error should have occurred due to invalid CaseSensitiveToken")

}

type Private struct {
	field string
}

type StructWithPrivateFields struct {
	Labels  map[string]string `label:"*"`
	private Private
	Public  string `label:"public"`
}

func TestIgnoreUnaccessibleFields(t *testing.T) {
	l := map[string]string{"public": "value"}
	priv := &StructWithPrivateFields{}
	err := Unmarshal(l, priv)
	assert.NoError(t, err)
	assert.Equal(t, l["public"], priv.Public)
}

// type WithValidation struct {
// 	Name          string            `label:"name"`
// 	Enum          MyEnum            `label:"enum,required"`
// 	RequiredField string            `label:"required_field,required"`
// 	Defaulted     string            `label:"defaulted,default:default value"`
// 	Labels        map[string]string `label:"*"`
// }

// func TestLabeleeWithValidation(t *testing.T) {
// 	l := StructWithLabels{
// 		Labels: map[string]string{
// 			"name": "my name",
// 			"enum": "X",
// 		},
// 	}
// 	v := &WithValidation{}
// 	err := Unmarshal(l, v)
// 	assert.Error(t, err, "should contain errors")
// 	var e *ParsingError
// 	if errors.As(err, &e) {
// 		assert.Len(t, e.Errors, 2)
// 	} else {
// 		assert.Fail(t, "error should be a parsing error")
// 	}
// 	assert.Equal(t, "my name", v.Name)
// 	assert.Equal(t, EnumUnknown, v.Enum)
// }

// type InvalidDueToMissingLabels struct {
// 	Name string `label:"name,required"`
// }

// type InvalidDueMyEnumErr struct {
// 	Enum MyEnum `label:"enum,required"`
// }

// func TestInvalidDueToMyEnumReturningError(t *testing.T) {
// 	l := StructWithLabels{
// 		Labels: map[string]string{
// 			"enum": "Invalid",
// 		},
// 	}

// 	inv := &InvalidDueMyEnumErr{}
// 	err := Unmarshal(l, inv)
// 	assert.Error(t, err, "Should have thrown an error")
// 	assert.Error(t, err)
// 	if !errors.Is(err, ErrParsing) {
// 		assert.Fail(t, "Error should be ErrInvalidValue")
// 	}
// 	var parsingError *ParsingError
// 	if errors.As(err, &parsingError) {
// 		assert.Equal(t, 1, len(parsingError.Errors))
// 		fieldErr := parsingError.Errors[0]
// 		fmt.Println(fieldErr)
// 		if !errors.Is(fieldErr, ErrExampleInvalidEnum) {
// 			assert.Fail(t, "Error should be ErrExampleInvalidEnum")
// 		} else {
// 			assert.Equal(t, "Enum", fieldErr.Field)
// 		}
// 	} else {
// 		assert.Fail(t, "Error should be a ParsingError")
// 	}
// 	t.Log(err)
// }

// type InvalidDueMultipleRequiredFields struct {
// 	Enum MyEnum `label:"enum,required"`
// 	Name string `label:"name,required"`
// }

// func TestInvalidDueToMultipleRequiredFields(t *testing.T) {
// 	l := StructWithLabels{
// 		Labels: map[string]string{},
// 	}

// 	inv := &InvalidDueMultipleRequiredFields{}
// 	err := Unmarshal(l, inv)
// 	assert.Error(t, err, "Should have thrown an error")
// 	assert.Error(t, err)

// 	var parsingError *ParsingError
// 	if errors.As(err, &parsingError) {
// 		assert.Equal(t, 2, len(parsingError.Errors))
// 	} else {
// 		assert.Fail(t, "Error should be a ParsingError")
// 	}

// 	t.Log(err)
// }
