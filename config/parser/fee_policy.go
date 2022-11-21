/*
 * Copyright 2022 LimeChain Ltd.
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

package parser

type FeePolicy struct {
	LegalEntities map[string]*LegalEntity `yaml:"policies,omitempty" json:"policies,omitempty"`
}

type LegalEntity struct {
	Addresses  []string   `yaml:"addresses,omitempty" json:"addresses,omitempty"`
	PolicyInfo PolicyInfo `yaml:"policy,omitempty" json:"policy,omitempty"`
}

type PolicyInfo struct {
	FeeType  string      `yaml:"fee_type,omitempty" json:"feeТype,omitempty"`
	Networks []uint64    `yaml:"networks,omitempty" json:"networks,omitempty"`
	Value    interface{} `yaml:"value,omitempty" json:"value,omitempty"`
}
