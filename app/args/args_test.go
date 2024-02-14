package args_test

import (
	"reflect"
	"testing"

	"github.com/aeron/digitalocean-ddns-updater/app/args"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgName(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name:     "normal-case",
			field:    reflect.StructField{},
			expected: "",
		},
		{
			name:     "field-has-tag",
			field:    reflect.StructField{Tag: reflect.StructTag(`name:"foo"`)},
			expected: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := args.Arg{Field: &tt.field}
			assert.Equal(t, tt.expected, arg.Name())
		})
	}
}

func TestArgFallback(t *testing.T) {
	tests := []struct {
		name       string
		field      reflect.StructField
		defaultVal string
		environVar string
		expected   string
	}{
		{
			name: "default-empty",
			field: reflect.StructField{
				Name: "foo",
				Type: reflect.PointerTo(reflect.TypeOf("")),
			},
			defaultVal: "",
			environVar: "",
			expected:   "",
		},
		{
			name: "default-not-empty",
			field: reflect.StructField{
				Name: "moo",
				Type: reflect.PointerTo(reflect.TypeOf("")),
			},
			defaultVal: "bar",
			environVar: "",
			expected:   "bar",
		},
		{
			name: "env-not-empty",
			field: reflect.StructField{
				Name: "goo",
				Type: reflect.PointerTo(reflect.TypeOf("")),
			},
			defaultVal: "rab",
			environVar: "VAR",
			expected:   "bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.environVar != "" {
				t.Setenv(tt.environVar, tt.expected)
			}

			arg := args.Arg{
				Field:   &tt.field,
				Default: tt.defaultVal,
				Environ: tt.environVar,
			}
			value, err := arg.Fallback()

			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestArgFlag(t *testing.T) {
	tests := []struct {
		name       string
		field      reflect.StructField
		defaultVal any
		kind       reflect.Kind
	}{
		{
			name: "int",
			field: reflect.StructField{
				Name: "foo",
				Type: reflect.PointerTo(reflect.TypeOf(5)),
			},
			defaultVal: 5,
			kind:       reflect.Int,
		},
		{
			name: "float64",
			field: reflect.StructField{
				Name: "moo",
				Type: reflect.PointerTo(reflect.TypeOf(5.0)),
			},
			defaultVal: 5.0,
			kind:       reflect.Float64,
		},
		{
			name: "string",
			field: reflect.StructField{
				Name: "goo",
				Type: reflect.PointerTo(reflect.TypeOf("bar")),
			},
			defaultVal: "bar",
			kind:       reflect.String,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := args.Arg{Field: &tt.field}
			val := arg.Flag(tt.defaultVal)

			require.NotNil(t, val)
			assert.Equal(t, tt.kind, reflect.ValueOf(val).Elem().Kind())
		})
	}
}

func TestParseSuccess(t *testing.T) {
	t.Setenv("VAR", "11")

	as := struct {
		A *int     `default:"10" environ:"VAR"`
		B *float64 `default:".5"`
		C *string  `default:"test"`
	}{}
	err := args.Parse(&as)

	require.NoError(t, err)
	assert.Equal(t, 11, *as.A)
	assert.Equal(t, .5, *as.B)
	assert.Equal(t, "test", *as.C)
}

func TestParseFailWithEnvVarType(t *testing.T) {
	t.Setenv("VAR", "not-an-int")

	as := struct {
		D *int `environ:"VAR"`
	}{}
	err := args.Parse(&as)

	require.Error(t, err)
}

func TestParseFailWithNonPtr(t *testing.T) {
	as := struct {
		E *int `default:"10"`
	}{}
	err := args.Parse(as)

	require.Error(t, err)
}

func TestParseFailWithWrongPtr(t *testing.T) {
	as := 125
	err := args.Parse(&as)

	require.Error(t, err)
}

func TestParseFailWithNonPtrField(t *testing.T) {
	as := struct {
		E int `default:"10"`
	}{}
	err := args.Parse(&as)

	require.Error(t, err)
}
