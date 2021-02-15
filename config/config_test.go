package config

import (
	"testing"

	"reflect"
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
