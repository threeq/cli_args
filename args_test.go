package args

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type TestArg1 struct {
	Name     string `yaml:"name" json:"name" toml:"name"`
	Arg      int    `yaml:"arg" json:"arg" toml:"arg"`
	Ignore   int
	InnerArg *TestInnerArg `yaml:"inner" json:"inner" toml:"inner"`
}

type TestInnerArg struct {
	Name string `yaml:"name" json:"name" toml:"name"`
	Arg  int    `yaml:"arg" json:"arg" toml:"arg"`
	Age  uint8    `yaml:"age" json:"age" toml:"age"`
	Array []string `yaml:"array" json:"array" toml:"array"`
	Map map[string]string `yaml:"map" json:"map" toml:"map"`
}

func Test_Bean2Args(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want map[string]*StructArg
	}{
		{"test", &TestArg1{InnerArg: &TestInnerArg{}}, map[string]*StructArg{
			"name":       {Name: "name", TName: "string", T: reflect.String},
			"arg":        {Name: "arg", TName: "int", T: reflect.Int},
			"inner.name": {Name: "inner.name", TName: "string", T: reflect.String},
			"inner.arg":  {Name: "inner.arg", TName: "int", T: reflect.Int},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Bean2Args(tt.args)
			for k, expect := range tt.want {
				actual := got[k]
				actual.Set = nil
				if fmt.Sprintf("%+v", actual) != fmt.Sprintf("%+v", expect) {
					t.Errorf("Bean2Args() = %v, want %v", fmt.Sprintf("%+v", actual), fmt.Sprintf("%+v", expect))
				}
			}
		})
	}
}

func Test_Bean2ArgsNotSupport(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want map[string]*StructArg
	}{
		{"test", &TestArg1{InnerArg: &TestInnerArg{}}, map[string]*StructArg{
			"inner.array":  nil,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Bean2Args(tt.args)
			for k, expect := range tt.want {
				actual := got[k]
				if actual != expect {
					t.Errorf("Bean2Args() = %v, want %v", actual, expect)
				}
			}
		})
	}
}

func Test_Usage(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app", "-h"}
	appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "test-default.yaml", false, ""))
	err := appArgs.Run(args)
	assert.Equal(t, err, ErrHelp)

	args = []string{"test-app", "-help"}
	appArgs = New("", Store(testCfg), FileConfigEnabled("config", "test-default.yaml", false, ""))
	err = appArgs.Run(args)
	assert.Equal(t, err, ErrHelp)
}
