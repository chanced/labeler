package labeler

import "reflect"

type setter func(r reflected, s string, o Options) error
type isetter = func(r reflected) setter

type fieldSetter func(f *field, s string, o Options) error
type pkgSetters = map[string]typeFieldSetters
type typeFieldSetters = map[string]fieldSetter

var pkgFieldSetters pkgSetters = pkgSetters{
	"time": typeFieldSetters{
		"Time": func(f *field, s string, o Options) error {
			return f.setTime(s, o)
		},
		"Duration": func(f *field, s string, o Options) error {
			return f.setDuration(s, o)
		},
	},
}

var basicFieldSetters = map[reflect.Kind]fieldSetter{
	reflect.Bool: func(f *field, s string, o Options) error {
		return f.setBool(s, o)
	},
	reflect.Float64: func(f *field, s string, o Options) error {
		return f.setFloat(s, 64, o)
	},
	reflect.Float32: func(f *field, s string, o Options) error {
		return f.setFloat(s, 32, o)
	},
	reflect.Int: func(f *field, s string, o Options) error {
		return f.setInt(s, 0, o)
	},
	reflect.Int8: func(f *field, s string, o Options) error {
		return f.setInt(s, 8, o)
	},
	reflect.Int16: func(f *field, s string, o Options) error {
		return f.setInt(s, 16, o)
	},
	reflect.Int32: func(f *field, s string, o Options) error {
		return f.setInt(s, 32, o)
	},
	reflect.Int64: func(f *field, s string, o Options) error {
		return f.setInt(s, 64, o)
	},
	reflect.String: func(f *field, s string, o Options) error {
		return f.setString(s, o)
	},
	reflect.Uint: func(f *field, s string, o Options) error {
		return f.setUint(s, 0, o)
	},
	reflect.Uint8: func(f *field, s string, o Options) error {
		return f.setUint(s, 8, o)
	},
	reflect.Uint16: func(f *field, s string, o Options) error {
		return f.setUint(s, 16, o)
	},
	reflect.Uint32: func(f *field, s string, o Options) error {
		return f.setUint(s, 32, o)
	},
	reflect.Uint64: func(f *field, s string, o Options) error {
		return f.setUint(s, 64, o)
	},
	reflect.Complex64: func(f *field, s string, o Options) error {
		return f.setComplex(s, 64, o)
	},
	reflect.Complex128: func(f *field, s string, o Options) error {
		return f.setComplex(s, 128, o)
	},
}
