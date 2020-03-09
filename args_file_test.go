package args

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CmdArgParseError(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app", "-config=zzz", "-c"}
	appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "test-default.yaml", false, ""))

	err := appArgs.Run(args[1:])
	assert.Equal(t, ErrCmdParse, err)
}

func Test_ConfigFileError(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}

	func() {
		args := []string{"test-app", "-config=test_data/not-found.yaml"}
		appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "", true, ""))
		err := appArgs.Run(args)
		assert.Equal(t, ErrFileNotFound, err)
	}()

	func() {
		args := []string{"test-app", "-config=test_data/error-format.json"}
		appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "", true, ""))
		err := appArgs.Run(args)
		assert.Equal(t, ErrFileParse, err)
	}()

	func() {
		args := []string{"test-app"}
		appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "", true, ""))
		err := appArgs.Run(args)
		assert.NotNil(t, err)
		assert.Equal(t, "flag required but not provided: -config", err.Error())
	}()

	func() {
		args := []string{"test-app", "-config=test_data/error-type.xxx"}
		appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "", true, ""))
		err := appArgs.Run(args)
		assert.Equal(t, ErrFileType, err)
	}()
}

func Test_ConfigFileArgDefault(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app"}
	appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "test_data/test-default.yaml", false, ""))

	err := appArgs.Run(args)
	assert.Nil(t, err)

	assert.Equal(t, testCfg.Name, "test-default-name")
	assert.Equal(t, testCfg.Arg, 100)
	assert.Equal(t, testCfg.InnerArg.Name, "test-default-inner-name")
	assert.Equal(t, testCfg.InnerArg.Arg, 200)
}

func Test_ConfigFileArg(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app", "-config=test_data/test.yaml"}
	appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "test-default.yaml", false, ""))

	err := appArgs.Run(args)
	assert.Nil(t, err)

	assert.Equal(t, testCfg.Name, "test-name")
	assert.Equal(t, testCfg.Arg, 1)
	assert.Equal(t, testCfg.InnerArg.Name, "test-inner-name")
	assert.Equal(t, testCfg.InnerArg.Arg, 11)
}

func Test_ConfigFileJsonFormat(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app", "-config=test_data/test.json"}
	appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "test-default.yaml", false, ""))

	err := appArgs.Run(args)
	assert.Nil(t, err)

	assert.Equal(t, "test-name", testCfg.Name)
	assert.Equal(t, 22, testCfg.Arg)
	assert.Equal(t, "test-inner-name", testCfg.InnerArg.Name)
	assert.Equal(t, 222, testCfg.InnerArg.Arg)
}

func Test_ConfigFileTomlFormat(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app", "-config=test_data/test.toml"}
	appArgs := New(args[0], Store(testCfg), FileConfigEnabled("config", "", true, ""))

	err := appArgs.Run(args)
	assert.Nil(t, err)

	assert.Equal(t, "test-name", testCfg.Name)
	assert.Equal(t, 33, testCfg.Arg)
	assert.Equal(t, "test-inner-name", testCfg.InnerArg.Name)
	assert.Equal(t, 333, testCfg.InnerArg.Arg)
}
