# labeler 🏷️

## ⚠️⚠️⚠️ Not ready for usage yet ⚠️⚠️⚠️

A Go package for marshaling and unmarshaling `map[string]string` with struct tags.

```bash
go get github.com/chanced/labeler
```

- [Value: `interface{}` defined](#value-interface-defined)
  - [Fields](#fields)
  - [Labels](#labels)
- [Input (Unmarshal)](#input-unmarshal)
- [Labeler Instance](#labeler-instance)
- [Examples](#examples)
  - [Basic with accessor / mutator for labels](#basic-example-with-accessor-mutator-for-labels)
  - [With an enum](#example-with-an-enum)
  - [Using a container tag](#example-using-a-container-tag)
  - [Using multiple tags](#example-using-multiple-tags)
- [Options](#options)
  - [Settings](#settings)
  - [Tokens](#tokens)
- [Errors](#errors)
  - [Value](#value-errors)
  - [Input](#input-errors)
  - [Field](#field-errors)
- [Notes](#notes)
  - [Comments](#comments)
  - [Prior Art](#prior-art)
  - [Feedback](#feedback)
- [License (MIT)](#license)

## Value: `interface{}` defined

Both Marshal and Unmarshal accept `v interface{}`, the value to marshal from or unmarshal
into. `v` must be pointer to a `struct` or a `type` that implements
`labeler.MarshalWithOpts`, `labeler.Marshal`, `labeler.UnmarshalWithOpts`, or
`labeler.Unmarshal` respectively.

### Fields

labeler is fairly flexible when it comes to what all you can tag. It supports the following types:

| Interface / Type              | Signature                                                                                                                                                   |     Usage |
| :---------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------- | --------: |
| `labeler.MarshalerWithOpts`   | `MarshalLabels(o labeler.Options) (map[string]string, error)`                                                                                               |   Marshal |
| `labeler.Marshaler`           | `MarshalLabels() (map[string]string, error)`                                                                                                                |   Marshal |
| `fmt.Stringer`                | `String() string`                                                                                                                                           |   Marshal |
| `encoding.TextMarshaler`      | `MarshalText() (text []byte, err error)`                                                                                                                    |   Marshal |
| `labeler.UnmarshalerWithOpts` | `UnmarshalLabels(v map[string]string, opts Options) error`                                                                                                  | Unmarshal |
| `labeler.Unmarshaler`         | `UnmarshalLabels(l map[string]string) error`                                                                                                                | Unmarshal |
| `Stringee`                    | `FromString(s string) error`                                                                                                                                | Unmarshal |
| `encoding.TextUnmarshaler`    | `UnmarshalText(text []byte) error`                                                                                                                          | Unmarshal |
| `struct`                      | can either implement any of the above interfaces or have fields with tags. Supports `n` level of nesting                                                    |      Both |
| basic types                   | `string`, `bool`, `int`, `int64`, `int32`, `int16`, `int8`, `float64`, `float32`, `uint`, `uint64`, `uint32`, `uint16`, `uint8`, `complex128`, `complex64`, |      Both |
| time                          | `time.Time`, `time.Duration`                                                                                                                                |      Both |
| pointer                       | pointer to any of the above                                                                                                                                 |      Both |

### Labels

When unmarshaling, labeler needs a way to persist labels, regardless of whether or not they
have been assigned to tagged fields. By default, labeler will retain all labels unless
`Options.KeepLabels` is set to `false` (See [Options](#options)). This is to ensure
data integrity for labels that have not been unmarshaled into fields.

Your available choices for `v` are:

| Interface / Type                | Signature                                                                                                                                                                                                                                                     |
| :------------------------------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `labeler.UnmarshalerWithOpts`   | `UnmarshalLabels(v map[string]string, opts Options) error`                                                                                                                                                                                                    |
| `labeler.Unmarshaler`           | `UnmarshalLabels(l map[string]string) error`                                                                                                                                                                                                                  |
| `labeler.Labelee`               | `SetLabels(labels map[string]string)`                                                                                                                                                                                                                         |
| `labeler.StrictLabelee`         | `SetLabels(labels map[string]string) error`                                                                                                                                                                                                                   |
| `labeler.GenericLabelee`        | `SetLabels(labels map[string]string, tag string) error`                                                                                                                                                                                                       |
| `struct` with a container field | A `struct` with a field marked as being the container using `Options.ContainerToken` on any level of `v` or a field with the name matching `Options.ContainerField` that is any `type` above or an accessible `map[string]string`. (See [Options](#options)). |

While marshaling, labeler will prioritize field values over those stored in your label
container. This means that values in the `map[string]string` will be overridden
if there is a key with a matching tag.

labeler ignores the case of keys by default, but this is configurable. (See [Options](#options))

## Input (Unmarshal)

For `Unmarshal`, you also need to pass `input interface{}` which provides a means of
accessing the labels `map[string]string`.

`input` must satisfy one of the following:

| Interface / Type             | Signature                                 | Example                                                    |
| :--------------------------- | :---------------------------------------- | ---------------------------------------------------------- |
| `labeler.Labeled`            | `GetLabels() map[string]string`           | [example](#basic-example-with-accessor-mutator-for-labels) |
| `labeler.GenericallyLabeled` | `GetLabels(tag string) map[string]string` | [example](#example-using-multiple-tags)                    |
| `map[string]string`          | `map[string]string`                       | [example](#example-using-a-container-tag)                  |

## Examples

### Basic example with accessor / mutator for labels

```go
package main

import (
    "fmt"

    "github.com/chanced/labeler"
)

type ExampleInput struct {
	Labels map[string]string
}

func (e ExampleInput) GetLabels() map[string]string {
	return e.Labels
}

type NestedExample struct {
	Field string `label:"nested_field"`
}

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
		"dedupe":          "Will be removed from the Labels after field value is set",
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
                    // something is wrong with the tag
            //  case ...
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

    fmt.Println(l) // map[string]string
}
```

### Example using a container tag

If you don't want to use accessors and mutators on your input and value, you can opt instead for a container tag that is of `map[string]string` or implements the appropriate `interface` to retrieve and set labels. You can also pass in a `map[string]string` directly as your input for Unmarshal.

The container token is configurable with the `ContainerToken` option. See [Options](#options) for more info.

```go
package main

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

    err := labeler.Unmarshal(l, v)
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

### Example with an enum

The only important bit is that `Color` implements `String() string` and
`FromString(s string) error`. `Color` could have also implemented
`UnmarshalText(text []byte) error` and `MarshalText() (text []byte, err error)`.

```go
package main

type Color int

const (
	ColorUnknown Color = iota
	ColorRed
	ColorBlue
)

type Example3 struct {
	Color  Color             `label:"color"`
	Labels map[string]string `label:"*"`
}

var colorMapToStr map[Color]string = map[Color]string{
	ColorUnknown: "Unknown",
	ColorBlue:    "Blue",
	ColorRed:     "Red",
}


func getColorMapFromStr() map[string]Color {
	m := make(map[string]Color)
	for k, v := range colorMapToStr {
		m[v] = k
	}
	return m
}

var colormMapFromStr map[string]Color = getColorMapFromStr()

func (c Color) String() string {
	return colorMapToStr[my]
}

func (c *Color) FromString(s string) error {
	if v, ok := colorMapFromStr[s]; ok {
		*my = v
		return nil
	}
	return errors.New("Invalid value")
}

type Example3 struct {
}

```

### Example using multiple tags

Say you have multiple sources of labels and you want to unmarshal them into the same `struct`. This is achievable by setting the option `Tag`! See [Options](#options) for more info.

```go
import (

    "github.com/chanced/labeler"
)

type Example4 struct {
    Name             string            `property:"name"`
    Color            string            `attribute:"color"`
    Characteristics  map[string]string
    Attributes       map[string]string
}

func (e *Example4) GetLabels(t string){
    switch t {
        case "property":
            return e.Characteristics
        case "attribute":
            return e.Attributes
    }
}
func (e *Example4) SetLabels(l map[string]string, t string) error{
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
    v := &Example4{}

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

### Settings

| Option                |  Default  | Details                                                                                                                                                                                                                                                                                                                                                                                                               | Option `func`                          |
| :-------------------- | :-------: | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :------------------------------------- |
| `Tag`                 | `"label"` | `Tag` is the name of the tag to lookup. This is especially handy if you have multiple sources of labels                                                                                                                                                                                                                                                                                                               | `OptTag(t string)`                     |
| `Seperator`           |   `","`   | Seperates the tag attributes. Configurable incase you have a tag that contains commas.                                                                                                                                                                                                                                                                                                                                | `OptSeperator(v string)`               |
| `ContainerField`      |   `""`    | `ContainerField` determines the field to set and retrieve the labels in the form of `map[string]string`. If `ContainerField` is set, labeler will assume that `GetLabels` and `SetLabels` should not be utilized. To set the `ContainerField` of a nested field, use dot notation (`Root.Labels`). <br>`ContainerField` is not required if `input` implements the appropriate `interface` to retrieve and set labels. | `OptContainerField(s string)`          |
| `ContainerToken`      |   `"*"`   | Used to indicate the field containing the `map[string]string` if you do not want to implement mutator/accessor interfaces. This can also be used to set some labels (see last column)                                                                                                                                                                                                                                 | `OptContainerToken(v string)`          |
| `AssignmentStr`       |   `":"`   | Used to assign values. This is in the event that a default value needs to contain `":"`                                                                                                                                                                                                                                                                                                                               | `OptAssignmentStr(v string)`           |
| `RequireAllFields`    |  `false`  | If `true`, all fields are required by default. An `ParsingError` will be returned from `Unmarshal` containing a `FieldError` for each field that was not present in the labels                                                                                                                                                                                                                                        | `OptRequireAllFields()`                |
| `KeepLabels`          |  `true`   | Indicates whether or not labels that have been assigned to values are kept in the labels `map[string]string` when unmarshaling.                                                                                                                                                                                                                                                                                       | `OptKeepLabels()` `OptDiscardLabels()` |
| `IgnoreCase`          |  `true`   | If `true`, label keys are matched regardless of case. Setting this to `false` makes all keys case sensitive. This can be overriden at the field level.                                                                                                                                                                                                                                                                | `OptCaseSensitive()`                   |
| `OmitEmpty`           |  `true`   | Determines whether or not to set zero-value when marshaling and unmarshaling.                                                                                                                                                                                                                                                                                                                                         | `OptOmitEmpty()` `OptIncludeEmpty()`   |
| `TimeFormat`          |   `""`    | Default format / layout to use when formatting `time.Time`. Field level formats can be provided with either `format` (configurable) or `timeformat` (configurable)                                                                                                                                                                                                                                                    | `OptTimeFormat(v string)`              |
| `IntBase`             |   `10`    | default base while parsing `int`, `int64`, `int32`, `int16`, `int8`                                                                                                                                                                                                                                                                                                                                                   | `OptIntBase(b int)`                    |
| `UintBase`            |   `10`    | default base while parsing `uint`, `uint64`, `uint32`, `uint16`, `uint8`                                                                                                                                                                                                                                                                                                                                              | `OptUintBase(b int)`                   |
| `ComplexFormat`       |   `'f'`   | Default format to use when formatting `complex` values.                                                                                                                                                                                                                                                                                                                                                               | `OptComplexFormat(f byte)`             |
| `FloatFormat`         |   `'f'`   | Default format to use when formatting `float` values.                                                                                                                                                                                                                                                                                                                                                                 | `OptFloatFormat(f byte)`               |
| `CaseSensitiveTokens` |  `true`   | Determines whether or not tokens, such as `required` and `ignorecase`, are case sensitive. This does not affect whether a label's key is case sensitive.                                                                                                                                                                                                                                                              | `OptCaseSensitiveTokens(v bool)`       |

### Tokens

| Option               |      Default      | Details                                                                                                                                           | Option `func`                     |
| :------------------- | :---------------: | :------------------------------------------------------------------------------------------------------------------------------------------------ | :-------------------------------- |
| `FormatToken`        |    `"format"`     | Token to set the field-level formatting for `time` and `float`.                                                                                   | `OptFormatToken(v string)`        |
| `ComplexFormatToken` | `"complexformat"` | Token used to set `ComplexFormat`. `FormatToken` can be used on non-container fields instead.                                                     | `OptComplexFormatToken(v string)` |
| `FloatFormatToken`   |  `"floatformat"`  | Token used to set `FloatFormat`. `FormatToken` can be used on non-container fields instead.                                                       | `OptFloatFormatToken(v string)`   |
| `TimeFormatToken`    |  `"timeformat"`   | Token used to set `TimeFormat`. `FormatToken` can be used on non-container fields instead.                                                        | `OptTimeFormatToken(v string)`    |
| `RequiredToken`      |   `"required"`    | Token to mark a field as required. If applied to a container, all fields are required unless `NotRequiredToken` is present.                       | `OptRequiredToken(v string)`      |
| `NotRequiredToken`   |  `"notrequired"`  | Token to mark a field as not required. Only relevant if `RequireAllFields` has been set to `true`.                                                | `OptRequiredToken(v string)`      |
| `DefaultToken`       |    `"default"`    | Token to provide a default value if one is not set.                                                                                               | `OptDefaultToken(v string)`       |
| `BaseToken`          |     `"base"`      | sets the token for parsing base of `int`, `int64`, `int32`, `int16`, `int8`, `uint`, `uint64`, `uint32`, `uint16`, and `uint8` at the field level | `OptBaseToken(v string)`          |
| `IntBaseToken`       |    `"intbase"`    | sets the token for parsing base of `int`, `int64`, `int32`, `int16`, `int8`, at the container or field level                                      | `OptIntBaseToken(v string)`       |
| `UintBaseToken`      |   `"uintbase"`    | sets the token for parsing base of `uint`, `uint64`, `uint32`, `uint16`, `uint8`, at the container or field level                                 | `OptUintBaseToken(v string)`      |
| `CaseSensitiveToken` | `"casesensitive"` | Token used to set `IgnoreCase` to `false`                                                                                                         | `OptCaseSensitiveToken(v string)` |
| `IgnoreCaseToken`    |  `"ignorecase"`   | Token used to determine whether or not to ignore case of the field's (or all fields if on container) key                                          | `OptIgnoreCaseToken(v string)`    |
| `OmitEmptyToken`     |   `"omitempty"`   | Token used to determine whether or not to assign empty / zero-value labels                                                                        | `OptOmitEmptyToken(v string)`     |
| `IncludeEmptyToken`  | `"includeempty"`  | Token used to determine whether or not to assign empty / zero-value labels                                                                        | `OptIncludeEmptyToken(v string)`  |
| `KeepToken`          |     `"keep"`      | Token used to set `KeepLabels` to `true`                                                                                                          | `OptKeepToken(v string)`          |
| `DiscardToken`       |    `"discard"`    | Token used to set `KeepLabels` to `false`                                                                                                         | `OptDiscardToken(v string)`       |

## Errors

## Notes

### Comments

This is my first project in Go. I have no doubt that there are places that could use improvement
and that the project, as a whole, could have been written more efficiently and cleaner overall.

Having said that, I'm blow away by how performant Go is. Under the hood, each `struct` and each
of its fields are analyzed and processed in goroutines, passing the results and/or errors through
channels.

#### Motivation

I'm building a project on Google Cloud Platform and the resources can have labels. I'm using the
grpc clients so the response objects have a`GetLabels() map[string]string` method. This package
should be incredibly handy for anyone else on GCP. It is generic enough to be utilized in other
circumstances as well though.

### Prior Art

- [go-env](https://github.com/Netflix/go-env) by Netflix. This is the only package that I looked at that does something similar. It was a huge help in getting started.

### Feedback

I'd love your feedback. Feel free to shoot me an email (chanceusc@gmail.com) or submit a ticket.
Either way, I'd greatly appreciate it.

If you run into any issues or have any questions, please do submit a ticket.

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Do with it as you wish.
