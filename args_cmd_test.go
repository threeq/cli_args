package args

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCmd(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{
		"test-app",
		"-name=test-cmd-name",
		"-inner.name=test-cmd-inner-name",
		"-inner.arg=555",
	}

	appArgs := New(args[0], Store(testCfg))

	err := appArgs.Run(args)
	assert.Nil(t, err)
	assert.Equal(t, "test-cmd-name", testCfg.Name)
	assert.Equal(t, 0, testCfg.Arg)
	assert.Equal(t, "test-cmd-inner-name", testCfg.InnerArg.Name)
	assert.Equal(t, 555, testCfg.InnerArg.Arg)
}

func TestCmdOverride(t *testing.T) {
	testCfg := &TestArg1{InnerArg: &TestInnerArg{}}
	args := []string{
		"test-app",
		"-config=test_data/test.yaml",
		"-arg=5656",
		"-inner.arg=5555",
	}
	_ = os.Setenv("TEST_INNER_NAME", "test-cmd-env-inner-name")
	_ = os.Setenv("TEST_INNER_ARG", "4444")
	appArgs := New(args[0], Store(testCfg),
		FileConfigEnabled("config", "", true, ""),
		EnvArg("TEST"),
	)

	err := appArgs.Run(args)
	assert.Nil(t, err)
	assert.Equal(t, "test-name", testCfg.Name)
	assert.Equal(t, 5656, testCfg.Arg)
	assert.Equal(t, "test-cmd-env-inner-name", testCfg.InnerArg.Name)
	assert.Equal(t, 5555, testCfg.InnerArg.Arg)

	_ = os.Unsetenv("TEST_INNER_NAME")
	_ = os.Unsetenv("TEST_INNER_ARG")
}
