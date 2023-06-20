package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getEnvValueOrDefault_noEnvSet_useDefault(t *testing.T) {
	const (
		testKey    = "test_key"
		defaultVal = "defaultVal"
	)
	_ = os.Unsetenv(testKey)
	val := getEnvValueOrDefault(testKey, defaultVal)
	assert.Equal(t, defaultVal, val)
}

func Test_getEnvValueOrDefault_hasEnvSet_getEnvVal(t *testing.T) {
	const (
		testKey    = "test_key"
		testEnvVal = "test_val"
		defaultVal = "defaultVal"
	)
	assert.NoError(t, os.Setenv(testKey, testEnvVal))
	t.Cleanup(func() {
		_ = os.Unsetenv(testKey)
	})

	val := getEnvValueOrDefault(testKey, defaultVal)
	assert.Equal(t, testEnvVal, val)
}
