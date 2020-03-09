package args

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestEnv(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app"}
	appArgs := New(args[0], Store(testCfg))
	assert.Nil(t, os.Setenv("NAME", "test-name"))
	assert.Nil(t, os.Setenv("ARG", "1"))
	assert.Nil(t, os.Setenv("INNER_NAME", "test-inner-name"))
	err := appArgs.Run(args)

	assert.Nil(t, err)
	assert.Equal(t, "test-name", testCfg.Name)
	assert.Equal(t, 1, testCfg.Arg)
	assert.Equal(t, "test-inner-name", testCfg.InnerArg.Name)
	assert.Equal(t, 0, testCfg.InnerArg.Arg)

	_ = os.Unsetenv("NAME")
	_ = os.Unsetenv("ARG")
	_ = os.Unsetenv("INNER_NAME")
}

func TestEnvPrefix(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app"}
	appArgs := New(args[0], Store(testCfg), EnvArg("test"))
	assert.Nil(t, os.Setenv("NAME", "test-name"))
	assert.Nil(t, os.Setenv("TEST_ARG", "1"))
	assert.Nil(t, os.Setenv("TEST_INNER_NAME", "test-inner-name"))
	assert.Nil(t, os.Setenv("TEST_INNER_ARG", "333"))
	err := appArgs.Run(args)

	assert.Nil(t, err)
	assert.Equal(t, "", testCfg.Name)
	assert.Equal(t, 1, testCfg.Arg)
	assert.Equal(t, "test-inner-name", testCfg.InnerArg.Name)
	assert.Equal(t, 333, testCfg.InnerArg.Arg)
	_ = os.Unsetenv("NAME")
	_ = os.Unsetenv("TEST_ARG")
	_ = os.Unsetenv("TEST_INNER_NAME")
	_ = os.Unsetenv("TEST_INNER_ARG")
}

func TestEnvOverride(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{"test-app", "-config=test_data/test.yaml"}
	appArgs := New(args[0],
		Store(testCfg),
		FileConfigEnabled("config", "", false, ""))
	assert.Nil(t, os.Setenv("NAME", "test-env-name"))
	assert.Nil(t, os.Setenv("INNER_ARG", "444"))
	err := appArgs.Run(args)

	assert.Nil(t, err)
	assert.Equal(t, "test-env-name", testCfg.Name)
	assert.Equal(t, 1, testCfg.Arg)
	assert.Equal(t, "test-inner-name", testCfg.InnerArg.Name)
	assert.Equal(t, 444, testCfg.InnerArg.Arg)

	_ = os.Unsetenv("NAME")
	_ = os.Unsetenv("INNER_ARG")
}
