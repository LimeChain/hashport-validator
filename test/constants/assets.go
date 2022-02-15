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

package constants

import (
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
)

var (
	Networks = map[int64]*parser.Network{
		0: {
			Name: "Hedera",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					constants.Hbar: {
						Networks: map[int64]string{
							33: "0x0000000000000000000000000000000000000001",
						},
					},
				},
				Nft: nil,
			},
		},
		1: {
			Name: "Network1",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					"0xb083879B1e10C8476802016CB12cd2F25a896691": {
						Networks: map[int64]string{
							33: "0x0000000000000000000000000000000000000123",
						},
					},
				},
				Nft: nil,
			},
		},
		2: {
			Name: "Network2",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					"0x0000000000000000000000000000000000000000": {
						Networks: map[int64]string{
							0: "",
						},
					},
				},
				Nft: nil,
			},
		},
		3: {
			Name: "Network3",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					"0x0000000000000000000000000000000000000000": {
						Networks: map[int64]string{
							0: "",
						},
					},
				},
				Nft: nil,
			},
		},
		32: {
			Name: "Network32",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					"0x0000000000000000000000000000000000000000": {
						Networks: map[int64]string{
							0: "",
						},
					},
				},
				Nft: nil,
			},
		},
		33: {
			Name: "Network33",
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					"0x0000000000000000000000000000000000000000": {
						Networks: map[int64]string{
							0: constants.Hbar,
							1: "0xsome-other-eth-address",
						},
					},
				},
				Nft: nil,
			},
		},
	}
)
