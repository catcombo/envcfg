package envcfg

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type Cfg struct {
	BoolField   bool    `env:"BOOL_FIELD"`
	IntField    int     `env:"INT_FIELD"`
	UIntField   uint    `env:"UINT_FIELD"`
	FloatField  float32 `env:"FLOAT_FIELD"`
	StringField string  `env:"STRING_FIELD"`
}

var tempDir = os.TempDir()

func setupEnv() {
	os.Setenv("BOOL_FIELD", "true")
	os.Setenv("INT_FIELD", "-8")
	os.Setenv("UINT_FIELD", "7")
	os.Setenv("FLOAT_FIELD", "16.5")
	os.Setenv("STRING_FIELD", "32")
}

func setupEnvFile() *os.File {
	tmpfile, _ := ioutil.TempFile(tempDir, ".env")

	tmpfile.WriteString(`
BOOL_FIELD=true
 INT_FIELD = 4 
UINT_FIELD= 11
 # FLOAT_FIELD=8.25
STRING_FIELD=ABC`)

	return tmpfile
}

func assertErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Received unexpected error:\n%+v", err)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Not equal: \n"+
			"expected: %#v\n"+
			"actual  : %#v", expected, actual,
		)
	}
}

func TestFieldVisibility(t *testing.T) {
	type Cfg struct {
		BoolField   bool
		IntField    int     `env:""`
		floatField  float32 `env:"FLOAT_FIELD"`
		StringField string  `env:"STRING_FIELD"`
	}

	setupEnv()
	defer os.Clearenv()

	cfg := Cfg{}
	err := Load(&cfg)

	expected := Cfg{
		StringField: "32",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, cfg)
}

func TestOverrideFromEnv(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	cfg := Cfg{}
	err := Load(&cfg)

	expected := Cfg{
		BoolField:   true,
		IntField:    -8,
		UIntField:   7,
		FloatField:  16.5,
		StringField: "32",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, cfg)
}

func TestOverrideFromFile(t *testing.T) {
	tmpfile := setupEnvFile()
	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	cfg := Cfg{}
	err := LoadFile(tmpfile.Name(), &cfg)

	expected := Cfg{
		BoolField:   true,
		IntField:    4,
		UIntField:   11,
		FloatField:  0,
		StringField: "ABC",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, cfg)
}

func TestOverrideAll(t *testing.T) {
	setupEnv()
	defer os.Clearenv()

	tmpfile := setupEnvFile()
	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	cfg := Cfg{}
	err := LoadFile(tmpfile.Name(), &cfg)

	expected := Cfg{
		BoolField:   true,
		IntField:    -8,
		UIntField:   7,
		FloatField:  16.5,
		StringField: "32",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, cfg)
}

func TestNestedStruct(t *testing.T) {
	type Nested struct {
		Field int `env:"INT_FIELD"`
	}

	type Cfg struct {
		Field bool `env:"BOOL_FIELD"`
		More  Nested
	}

	setupEnv()
	defer os.Clearenv()

	cfg := Cfg{}
	err := Load(&cfg)

	expected := Cfg{
		Field: true,
		More: Nested{
			Field: -8,
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, cfg)
}
