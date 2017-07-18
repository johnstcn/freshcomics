package log


import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var debugBuf, infoBuf, errorBuf bytes.Buffer
var testMsg string = "test message:"

func TestDebug(t *testing.T) {
	testErr := errors.New("something happened")
	debugBuf.Reset()
	Debug(testMsg, testErr)
	result, err := debugBuf.ReadString('\n')
	assert.Nil(t, err)
	assert.NotEqual(t, "", result)
}

func TestInfo(t *testing.T) {
	testErr := errors.New("something happened")
	infoBuf.Reset()
	Info(testMsg, testErr)
	result, err := infoBuf.ReadString('\n')
	assert.Nil(t, err)
	assert.NotEqual(t, "", result)
}

func TestError(t *testing.T) {
	testErr := errors.New("something happened")
	errorBuf.Reset()
	Error(testMsg, testErr)
	result, err := errorBuf.ReadString('\n')
	assert.Nil(t, err)
	assert.NotEqual(t, "", result)
}

func TestMain(m *testing.M) {
	setupLogger(&debugBuf, &infoBuf, &errorBuf)
	os.Exit(m.Run())
}