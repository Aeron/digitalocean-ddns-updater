package args

import (
	"errors"
	"flag"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Represents a single argument
type Arg struct {
	Field   *reflect.StructField
	Default string
	Environ string
	Help    string
}

// Returns an argument name.
func (arg *Arg) Name() string {
	name := arg.Field.Tag.Get("name")
	if name == "" {
		return strings.ToLower(arg.Field.Name)
	}
	return name
}

// Returns a properly typed argument fallback value.
func (arg *Arg) Fallback() (any, error) {
	value := os.Getenv(arg.Environ)

	if value == "" {
		value = arg.Default
	}

	switch arg.Field.Type.Elem().Kind() {
	case reflect.Int:
		return strconv.Atoi(value)
	case reflect.Float64:
		return strconv.ParseFloat(value, 64)
	default:
		return value, nil
	}
}

// Returns a command-line flag value (pointer).
func (arg *Arg) Flag(fallback any) any {
	switch arg.Field.Type.Elem().Kind() {
	case reflect.Int:
		return flag.Int(arg.Name(), fallback.(int), arg.Help)
	case reflect.Float64:
		return flag.Float64(arg.Name(), fallback.(float64), arg.Help)
	case reflect.String:
		return flag.String(arg.Name(), fallback.(string), arg.Help)
	default:
		return nil
	}
}

// Returns a concluding argument value (pointer).
func (arg *Arg) Value() (any, error) {
	fallback, err := arg.Fallback()
	if err != nil {
		return nil, err
	}
	return arg.Flag(fallback), nil
}

// Parses arguments according to a given tagged structure.
func Parse(args any) error {
	ptr := reflect.ValueOf(args)
	val := ptr.Elem()
	typ := val.Type()

	if ptr.Kind() != reflect.Pointer || typ.Kind() != reflect.Struct {
		return errors.New("must be a pointer to a structure")
	}

	for _, field := range reflect.VisibleFields(typ) {
		arg := Arg{
			Field:   &field,
			Default: field.Tag.Get("default"),
			Environ: field.Tag.Get("environ"),
			Help:    field.Tag.Get("help"),
		}

		value, err := arg.Value()
		if err != nil {
			return err
		}

		if f := val.FieldByName(field.Name); f.CanSet() {
			f.Set(reflect.ValueOf(value))
		} else {
			return errors.New("cannot set an argument value")
		}
	}

	flag.Parse()

	return nil
}
