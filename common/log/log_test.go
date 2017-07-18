package log


import (
	"testing"
	"github.com/stretchr/testify/assert"
	"bytes"
	"io/ioutil"
)

var debugBuf, infoBuf, errorBuf *bytes.Buffer
var testMsg string = "test message"

func Test_Debug(t *testing.T) {
	expected := "[DEBUG] test message"
	debugBuf.Reset()
	Debug(testMsg)
	result, err := ioutil.ReadAll(debugBuf)
	assert.NotNil(t, err)
	assert.EqualValues(t, result, expected)
}

func Test_Info(t *testing.T) {
	expected := "[INFO] test message"
	infoBuf.Reset()
	Info(testMsg)
	result, err := ioutil.ReadAll(infoBuf)
	assert.NotNil(t, err)
	assert.EqualValues(t, result, expected)
}

func Test_Error(t *testing.T) {
	expected := "[ERROR] test message"
	errorBuf.Reset()
	Error(testMsg)
	result, err := ioutil.ReadAll(errorBuf)
	assert.NotNil(t, err)
	assert.EqualValues(t, result, expected)
}

func TestMain(m *testing.M) {
	setupLogger(debugBuf, infoBuf, errorBuf)
}