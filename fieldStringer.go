package labeler

import "reflect"

type fieldStringer func(f *field, o Options) (string, error)

func (get fieldStringer) Marshaler(r reflected, o Options) marshal {
	if r.Topic() != fieldTopic {
		return nil
	}
	return func(r reflected, kvs *keyvalues, o Options) error {
		f := r.(*field)
		s, err := get(f, o)
		if err != nil {
			return err
		}
		if s == "" {
			s = f.Default(o)
		}
		if o.OmitEmpty && s == "" {
			return nil
		}
		kvs.Set(f.Key, s)
		return nil
	}
}

var fieldStringerPkgs = map[string]map[string]fieldStringer{
	"time": {
		"Time": func(f *field, o Options) (string, error) {
			return f.formatTime(o)
		},
		"Duration": func(f *field, o Options) (string, error) {
			return f.formatDuration(o)
		},
	},
}

var fieldStringerBasic = map[reflect.Kind]fieldStringer{
	reflect.Bool: func(f *field, o Options) (string, error) {
		return f.formatBool(o)
	},
	reflect.Float64: func(f *field, o Options) (string, error) {
		return f.formatFloat(o)
	},
	reflect.Float32: func(f *field, o Options) (string, error) {
		return f.formatFloat(o)
	},
	reflect.Int: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int8: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int16: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int32: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.Int64: func(f *field, o Options) (string, error) {
		return f.formatInt(o)
	},
	reflect.String: func(f *field, o Options) (string, error) {
		return f.formatString(o)
	},
	reflect.Uint: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint8: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint16: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint32: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Uint64: func(f *field, o Options) (string, error) {
		return f.formatUint(o)
	},
	reflect.Complex64: func(f *field, o Options) (string, error) {
		return f.formatComplex(o)
	},
	reflect.Complex128: func(f *field, o Options) (string, error) {
		return f.formatComplex(o)
	},
}
