# labeler

!!!**Not ready for usage yet**

A Go package for marshaling and unmarshaling `map[string]string` with struct tags.

![myImage](https://media3.giphy.com/media/xT5LMx9pJT5Uvbs6D6/giphy.gif)

<sup>source: [giphy](https://giphy.com/gifs/season-3-the-simpsons-3x13-xT5LMx9pJT5Uvbs6D6)</sup>

```bash
go get github.com/chanced/labeler
```

## Value: `interface{}` defined

### Fields

labeler is fairly flexible when it comes to what all you can tag. It supports the following types:

| Interface / Type            | Signature                                                                                                 |        Usage |
| :-------------------------- | :-------------------------------------------------------------------------------------------------------- | -----------: |
| `MarshalerWithOptions`      | `MarshalLabels(o labeler.Options) (map[string]string, error)`                                             |   Marshaling |
| `MarshalerWithTagAndOpts`   | `MarshalLabels(t labeler.Tag, o labeler.Options) (map[string]string, error)`                              |   Marshaling |
| `Marshaler`                 | `MarshalLabels() (map[string]string, error)`                                                              |   Marshaling |
| `fmt.Stringer`              | `String() string`                                                                                         |   Marshaling |
| `encoding.TextMarshaler`    | `MarshalText() (text []byte, err error)`                                                                  |   Marshaling |
| `UnmarshalerWithTagAndOpts` | `UnmarshalLabels(v map[string]string, t Tag, opts Options) error`                                         | Unmarshaling |
| `UnmarshalerWithOptions`    | `UnmarshalLabels(v map[string]string, opts Options) error`                                                | Unmarshaling |
| `Unmarshaler`               | `UnmarshalLabels(l map[string]string) error`                                                              | Unmarshaling |
| `Stringee`                  | `FromString(s string) error`                                                                              | Unmarshaling |
| `encoding.TextUnmarshaler`  | `UnmarshalText(text []byte) error`                                                                        | Unmarshaling |
| `struct`                    | can either implement any of the above interfaces or have fields with tags. Supports `n` level of nesting  |         Both |
| basic types                 | `string`, `bool`, `int`, `int64`, `int32`, `int16`, `int8`, `uint`, `uint64`, `uint32`, `uint16`, `uint8` |         Both |
| time                        | `time.Time`, `time.Duration`                                                                              |         Both |

### Input

### Labels

When it comes to Unmarshaling, labeler needs a way to persist labels, regardless whether or not they have been assigned
to tagged fields. By default, labeler will retain all labels unless `Options.KeepLabels` is set to `false` (See [Options](#options)).

---

## Examples

### Basic example with accessor / mutator for Labels

```go
import (
    "github.com/chanced/labeler"
)

type ExampleInput struct {
	Labels map[string]string
}

func (e ExampleInput) GetLabels() map[string]string {
	return e.Labels
}
func (e *ExampleInput) SetLabels(l map[string]string) {
	e.Labels = l
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


type NestedExample struct {
	Field string `label:"nested_field"`
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
	Nested          NestedExample
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
		"name":            "Bart",
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
		"nested_field":    "nested value",
	}

	input := ExampleInput {
		Labels: labels,
	}
    v := &Example{}

    err := labeler.Unmarshal(input, v)

    if err != nil {
        var pErr *labeler.ParsingError
        if errors.As(err, &pErr){
            for _, fieldErr := range pErr.Errors {
                // fieldErr has the field's Name (string) and Tag (labeler.Tag)
                // as well as Err, the underlying Error which unwraps
                switch {
                    case errors.Is(fieldErr, ErrLabelRequired):
                        // a field marked as required is missing in the labels
                    case errors.Is(fieldErr, ErrMalformedTag):
                 // case ...
                }
            }
        } else if errors.Is(err, ErrInvalidInput) {
        }
        // see errors.go for all options
    }

    l, err := labeler.Unmarshal(v)

    if err != nil {
        //handle err
    }

    _ = l // map[string]string
}
```

### Example using a container tag

If you don't want to use accessors and mutators on your input and value, you can opt instead for a container tag that is of `map[string]string` or implements the appropriate `interface` to retrieve and set labels. You can also pass in a `map[string]string` directly as your input for Unmarshaling.

The container token is configurable with the `ContainerToken` option. See [Options](#options) for more info.

```go
import (
    "fmt"
    "github.com/chanced/labeler"
)

type Example2 struct {
	Name          string            `label:"name"`
	Defaulted     string            `label:"defaulted, default: my default value"` // spaces are trimmed
	Labels        map[string]string `label:"*, required"` // all fields are now required
}

func main() {
    l := map[string]string { name: "Homer" }

    v := &Example2{}

    err := labeler.Unmarshal(l, u)
    if err != nil {
        // handle err
        _ = err
    }
    fmt.Println("Name is: ", v.Name)
    fmt.Println(len(v.Labels), " Labels")
    labels, err := labeler.Marshal(v, l)
    if err != nil {
        _ = err
    }

}
```

### Example using multiple tags

Say you have multiple sources of labels and you want to unmarshal them into the same `struct`. This is achievable by setting the option `Tag`! See [Options](#options) for more info.

```go
import (

    "github.com/chanced/labeler"
)

type Example3 struct {
    Name             string            `property:"name"`
    Color            string            `attribute:"color"`
    Characteristics  map[string]string
    Attributes       map[string]string
}
// you could also use containers, which is probably easier but for demo purposes:
func (e *Example3) GetLabels(t string){
    switch t {
        case "property":
            return e.Characteristics
        case "attribute":
            return e.Attributes
    }
}
func (e *Example3) SetLabels(l map[string]string, t string) error{
    switch t {
        case "property":
            e.Characteristics = l
        case "attribute":
            e.Attributes = l
    }
}


func main() {
    properties := map[string]string { name: "Homer" }
    attributes := map[string]string { color: "Yellow" }
    v := &Example3{}

    err := labeler.Unmarshal(v, properties, OptTag("property"))
    if err != nil {
        _ = err
    }
   err := labeler.Unmarshal(v, attributes, OptTag("attribute"))
    if err != nil {
        _ = err
    }

}
```

## Options

| Option                |      Default      | Details                                                                                                                                                                                                                             | Option `func`                          |
| :-------------------- | :---------------: | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :------------------------------------- |
| `Tag`                 |     `"label"`     | `Tag` is the name of the tag to lookup. This is especially handy if you have multiple sources of labels                                                                                                                             | `OptTag(t string)`                     |
| `Seperator`           |       `","`       | Seperates the tag attributes. Configurable incase you have a tag that contains commas.                                                                                                                                              | `OptSeperator(v string)`               |
| `ContainerField`      |       `""`        | `ContainerField` is not required if `input` implements the appropriate `interface` to retrieve and set labels. If `ContainerField` is set, the `type` of `input` while Unmarshaling is not checked and `ContainerField` is instead. | `OptContainerField(s string)`          |
| `ContainerToken`      |       `"*"`       | Used to indicate the field containing the `map[string]string` if you do not want to implement mutator/accessor interfaces. This can also be used to set some labels (see last column)                                               | `OptContainerToken(v string)`          |
| `AssignmentStr`       |       `":"`       | Used to assign values. This is in the event that a default value needs to contain `":"`                                                                                                                                             | `OptAssignmentStr(v string)`           |
| `RequireAllFields`    |      `false`      | If `true`, all fields are required by default. An `ParsingError` will be returned from `Unmarshal` containing a `FieldError` for each field that was not present in the labels                                                      | `OptRequireAllFields()`                |
| `KeepLabels`          |      `true`       | Indicates whether or not labels that have been assigned to values are kept in the labels `map[string]string` when unmarshaling.                                                                                                     | `OptKeepLabels()` `OptDiscardLabels()` |
| `IgnoreCase`          |      `true`       | If `true`, label keys are matched regardless of case. Setting this to `false` makes all keys case sensitive. This can be overriden at the field level.                                                                              | `OptCaseSensitive()`                   |
| `TimeFormat`          |       `""`        | Default format / layout to use when formatting `time.Time`. Field level formats can be provided with either `format` (configurable) or `timeformat` (configurable)                                                                  | `OptTimeFormat(v string)`              |
| `FloatFormat`         |       `'f'`       | Default format to use when formatting `float` values.                                                                                                                                                                               | `OptFloatFormat(f byte)`               |
| `FormatToken`         |    `"format"`     | Token to set the field-level formatting for `time` and `float`.                                                                                                                                                                     | `OptFormatToken(v string)`             |
| `FloatFormatToken`    |  `"floatformat"`  | Token used to set `FloatFormat`. This is really only relevant on container fields as `FormatToken` can be used at the field level.                                                                                                  | `OptFloatFormatToken(v string)`        |
| `TimeFormatToken`     |  `"timeformat"`   | Token used to set `TimeFormat`. This is really only relevant on container fields as `FormatToken` can be used at the field level.                                                                                                   | `OptTimeFormatToken(v string)`         |
| `RequiredToken`       |   `"required"`    | Token to mark a field as required. If applied to a container, all fields are required unless `NotRequiredToken` is present.                                                                                                         | `OptRequiredToken(v string)`           |
| `NotRequiredToken`    |  `"notrequired"`  | Token to mark a field as not required. Only relevant if `RequireAllFields` has been set to `true`.                                                                                                                                  | `OptRequiredToken(v string)`           |
| `DefaultToken`        |    `"default"`    | Token to provide a default value if one is not set. Stylistic.                                                                                                                                                                      | `OptDefaultToken(v string)`            |
| `CaseSensitiveToken`  | `"casesensitive"` | Token used to set `IgnoreCase` to `false`                                                                                                                                                                                           | `OptCaseSensitiveToken(v string)`      |
| `IgnoreCaseToken`     |  `"ignorecase"`   | Token used to determine whether or not to ignore case of the field's (or all fields if on container) key                                                                                                                            | `OptIgnoreCaseToken(v string)`         |
| `KeepToken`           |     `"keep"`      | Token used to set `KeepLabels` to `true`                                                                                                                                                                                            | `OptKeepToken(v string)`               |
| `DiscardToken`        |    `"discard"`    | Token used to set `KeepLabels` to `false`                                                                                                                                                                                           | `OptDiscardToken(v string)`            |
| `CaseSensitiveTokens` |      `false`      | Determines whether or not tokens, such as `required` and `ignorecase`, are case sensitive. This does not affect whether a label's key is case sensitive.                                                                            | `OptCaseSensitiveTokens(v bool)`       |

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
