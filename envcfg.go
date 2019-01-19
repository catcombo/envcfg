// Package envcfg provides functions to load values to a structure fields from .env file and from OS environment variables.
//
// Usage
//
// Declare a structure and use tag `env` to define associated environment variable names for desired fields.
//
// 	type Cfg struct {
// 		Debug       bool   `env:"DEBUG"`
// 		DatabaseURL string `env:"DATABASE_URL"`
// 	}
//
// Create a new structure to provide default values.
//
//  cfg := Cfg{
//  	Debug: true,
//  	DatabaseURL: "sqlite:///db.sqlite",
//  }
//
// Call envcfg.Load() to load values from environment variables.
//
//  err := envcfg.Load(&cfg)
//
// Keep in mind that the values are first loaded from the .env file (if it exists) and then
// from the OS environment variables that can override the values loaded from the file.
//
// The syntax of the .env file should follow these rules:
//
//  - Each line should be in VAR=VAL format
//  - Lines beginning with # are processed as comments and ignored
//  - Blank lines are ignored
//
// Notes
//
// A limited number of field types are supported, but they should be enough for most cases.
// Nested structures are supported, just mark nested fields with `env` tag as usual
// (no special syntax for .env file required). To load .env file from a different location or
// with a different name use LoadFile() function.
package envcfg

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

type structField struct {
	Name  string
	Value *reflect.Value
}

type iterable interface {
	Iter() bool
	Next() string
}

type arrayIter struct {
	Index   int
	Environ *[]string
}

func newArrayIter(arr *[]string) *arrayIter {
	return &arrayIter{
		Index:   -1,
		Environ: arr,
	}
}

func (ai *arrayIter) Iter() bool {
	ai.Index += 1
	return ai.Index < len(*ai.Environ)
}

func (ai *arrayIter) Next() string {
	return (*ai.Environ)[ai.Index]
}

func loadFromEnv(fields []*structField) error {
	environ := os.Environ()
	return loadFromSource(newArrayIter(&environ), fields)
}

type scannerIter struct {
	Scanner *bufio.Scanner
}

func (si *scannerIter) Iter() bool {
	return si.Scanner.Scan()
}

func (si *scannerIter) Next() string {
	return si.Scanner.Text()
}

func loadFromFile(filename string, fields []*structField) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return loadFromSource(&scannerIter{
		Scanner: bufio.NewScanner(file),
	}, fields)
}

func setValue(field *reflect.Value, value string) error {
	k := field.Kind()

	switch {
	case k == reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case k >= reflect.Int && k <= reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case k >= reflect.Uint && k <= reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case k >= reflect.Float32 && k <= reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case k == reflect.String:
		field.SetString(value)
	default:
		return fmt.Errorf("field %s is not supported", field.Kind())
	}

	return nil
}

func readSource(source iterable) (*map[string]string, error) {
	vars := make(map[string]string)

	for source.Iter() {
		line := strings.TrimSpace(source.Next())

		// Skip blank lines and comments
		if (len(line) == 0) || strings.HasPrefix(line, "#") {
			continue
		}

		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) != 2 {
			return nil, fmt.Errorf("key and value must be separated by the sign '=': %v", line)
		}

		key := strings.TrimSpace(tokens[0])
		value := strings.TrimSpace(tokens[1])

		vars[key] = value
	}

	return &vars, nil
}

func loadFromSource(source iterable, fields []*structField) error {
	vars, err := readSource(source)
	if err != nil {
		return err
	}

	for _, field := range fields {
		if data, ok := (*vars)[field.Name]; ok {
			err := setValue(field.Value, data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func parseStruct(st *reflect.Value) []*structField {
	var fields []*structField
	t := st.Type()

	for i := 0; i < st.NumField(); i++ {
		refField := t.Field(i)
		refValue := st.Field(i)

		// Skip unexported field
		if refField.PkgPath != "" {
			continue
		}

		if refValue.Kind() == reflect.Struct {
			// Append fields from the nested struct
			fields = append(fields, parseStruct(&refValue)...)
		} else {
			// Skip field without tag
			envVarName := refField.Tag.Get("env")
			if envVarName == "" {
				continue
			}

			fields = append(fields, &structField{
				Name:  envVarName,
				Value: &refValue,
			})
		}
	}

	return fields
}

// Loads values from the specified file and OS environment variables to a structure
// passed by the reference to the function.
func LoadFile(filename string, to interface{}) error {
	value := reflect.ValueOf(to)

	// Loading target must be a pointer to structure
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		return errors.New("target must be a pointer to struct")
	}

	st := value.Elem()
	fields := parseStruct(&st)

	// Load environment variables from .env file
	err := loadFromFile(filename, fields)
	if err != nil {
		// Skip if file doesn't exist
		if _, statErr := os.Stat(filename); !os.IsNotExist(statErr) {
			return err
		}
	}

	// Override values by loading OS environment variables
	err = loadFromEnv(fields)
	if err != nil {
		return err
	}

	return nil
}

// Loads values from .env file located in the current working directory and
// from OS environment variables to a structure passed by the reference to the function.
func Load(to interface{}) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	filename := path.Join(dir, ".env")
	return LoadFile(filename, to)
}
