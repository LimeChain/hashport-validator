package config

import (
	"reflect"
	"testing"
)

func Test_LoadConfig(t *testing.T) {
	configuration := LoadConfig()
	if reflect.TypeOf(configuration).String() != "*config.Config" {
		t.Fatalf(`Expected to return configuration type *config.Config, but returned: [%s]`, reflect.TypeOf(configuration).String())
	}
}

func Test_LoadTestConfig(t *testing.T) {
	configuration := LoadTestConfig()
	if reflect.TypeOf(configuration).String() != "*config.Config" {
		t.Fatalf(`Expected to return configuration type *config.Config, but returned: [%s]`, reflect.TypeOf(configuration).String())
	}
}
func Test_GetConfig(t *testing.T) {
	var configuration Config

	// Test we get an error when wrong path is provided
	err := GetConfig(&configuration, "non-existing-path/application.yml")
	if err == nil {
		t.Fatalf(err.Error())
	}

	// Test we get no error when existing path is provided
	err = GetConfig(&configuration, mainConfigFile)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
