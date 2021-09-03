/*
 * Copyright 2021 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"reflect"
	"testing"
)

func Test_LoadConfig(t *testing.T) {
	configuration := LoadConfig()
	if reflect.TypeOf(configuration).String() != "config.Config" {
		t.Fatalf(`Expected to return configuration type *config.Config, but returned: [%s]`, reflect.TypeOf(configuration).String())
	}
}

func Test_GetConfig(t *testing.T) {
	var configuration Config

	// Test we get an error when wrong path is provided
	err := GetConfig(&configuration, "non-existing-path/application.yml")
	if err == nil {
		t.Fatalf(`Expected GetConfig to return error when loading non-existing application.yml file`)
	}

	// Test we get no error when existing path is provided
	err = GetConfig(&configuration, "node.yml")
	if err != nil {
		t.Fatalf(err.Error())
	}
}
