# labeler

A Go struct tags for marshaling and unmarshaling `map[string]string`

### Tagged fields can be

#### basic types:

`string`, `bool`, `int`, `int64`, `int32`, `int16`, `int8`, `uint`, `uint64`, `uint32`, `uint16`, `uint8`

#### Time:

`time.Time`, `time.Duration`

Or any type that implements:

| Interface                | Signature                                |     Used For |
| ------------------------ | :--------------------------------------- | -----------: |
| fmt.Stringer             | `String() string`                        |   Marshaling |
| labeler.Stringee         | `FromString(s string) error`             | Unmarshaling |
| encoding.TextMarshaler   | `MarshalText() (text []byte, err error)` |   Marshaling |
| encoding.TextUnmarshaler | `UnmarshalText(text []byte) error`       | Unmarshaling |

---

## Examples

### Basic example with accessor / mutator for Labels

```go
import (
    "github.com/chanced/labeler"
)
type Example struct {
	Name            string        `label:"name"`
	Important       string        `label:"imp, required"`
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
	Labels          map[string]string
}

func (e *Example) SetLabels(l map[string]string) {
	e.Labels = l
}

func (e *Example) GetLabels() map[string]string {
	return e.Labels
}

func main() {
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
        "dedupe":          "will not be in labels",
        "case":            "case sensitive label",
    }
    input := StructWithLabels{
        Labels: labels,
    }
    v := &Example{}

    err := labeler.Unmarshal(input, v)
    if err != nil {
        var pErr *labeler.ParsingError
        if errors.As(err, &pErr){
            for _, fieldErr := range pErr.Errors {
                switch {
                    case errors.Is(fieldErr, ErrLabelRequired):
                        //...
                    case errors.Is(fieldErr, ErrMalformedTag):
                        //...
                    default:
                        // see https://github.com/chanced/labeler/blob/master/errors.go
                }
            }
        } else if errors.Is(err, ErrInvalidInput) {
            //...
        }
    }

    labels, err := labeler.Unmarshal(v)

    if err != nil {
        //handle err
    }
}
```

### Example using a container tag

```go
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

```
